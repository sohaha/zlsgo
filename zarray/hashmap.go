//go:build go1.18
// +build go1.18

package zarray

import (
	"encoding/json"
	"reflect"
	"sort"
	"strconv"
	"sync/atomic"
	"unsafe"

	"github.com/sohaha/zlsgo/zutil"
	"golang.org/x/exp/constraints"
	"golang.org/x/sync/singleflight"
)

const (
	// defaultSize is the default size for a zero allocated map
	defaultSize = 8

	// maxFillRate is the maximum fill rate for the slice before a resize will happen
	maxFillRate = 50

	// intSizeBytes is the size in byte of an int or uint value
	intSizeBytes = strconv.IntSize >> 3
)

// indicates resizing operation status enums
const (
	notResizing uint32 = iota
	resizingInProgress
)

type (
	hashable interface {
		constraints.Integer | constraints.Float | constraints.Complex | ~string | uintptr | unsafe.Pointer
	}

	metadata[K hashable, V any] struct {
		count     *zutil.Uintptr
		data      unsafe.Pointer
		index     []*element[K, V]
		keyshifts uintptr
	}

	// Maper implements the concurrent hashmap
	Maper[K hashable, V any] struct {
		gsf         singleflight.Group
		listHead    *element[K, V]
		hasher      func(K) uintptr
		metadata    atomicPointer[metadata[K, V]]
		resizing    *zutil.Uint32
		numItems    *zutil.Uintptr
		defaultSize uintptr
	}

	deletionRequest[K hashable] struct {
		key     K
		keyHash uintptr
	}
)

func NewHashMap[K hashable, V any](size ...uintptr) *Maper[K, V] {
	m := &Maper[K, V]{
		listHead:    newListHead[K, V](),
		resizing:    zutil.NewUint32(0),
		numItems:    zutil.NewUintptr(0),
		defaultSize: defaultSize,
	}

	m.numItems.Store(0)
	if len(size) > 0 && size[0] != 0 {
		m.defaultSize = size[0]
	}

	m.allocate(defaultSize)
	m.setDefaultHasher()
	return m
}

// Delete a map data structure.
func (m *Maper[K, V]) Delete(keys ...K) {
	size := len(keys)
	switch {
	case size == 0:
		return
	case size == 1:
		var (
			h        = m.hasher(keys[0])
			existing = m.metadata.Load().indexElement(h)
		)
		if existing == nil || existing.keyHash > h {
			existing = m.listHead.next()
		}
		for ; existing != nil && existing.keyHash <= h; existing = existing.next() {
			if existing.key == keys[0] {
				if existing.remove() {
					m.removeItemFromIndex(existing)
				}
				return
			}
		}
	default:
		var (
			delQueue = make([]deletionRequest[K], size)
			iter     = 0
		)
		for idx := 0; idx < size; idx++ {
			delQueue[idx].keyHash, delQueue[idx].key = m.hasher(keys[idx]), keys[idx]
		}

		sort.Slice(delQueue, func(i, j int) bool {
			return delQueue[i].keyHash < delQueue[j].keyHash
		})

		elem := m.metadata.Load().indexElement(delQueue[0].keyHash)

		if elem == nil || elem.keyHash > delQueue[0].keyHash {
			elem = m.listHead.next()
		}

		for elem != nil && iter < size {
			if elem.keyHash == delQueue[iter].keyHash && elem.key == delQueue[iter].key {
				if elem.remove() {
					m.removeItemFromIndex(elem)
				}
				iter++
				elem = elem.next()
			} else if elem.keyHash > delQueue[iter].keyHash {
				iter++
			} else {
				elem = elem.next()
			}
		}
	}
}

// Has a parameter "key" of type K and returns a boolean value "ok".
func (m *Maper[K, V]) Has(key K) (ok bool) {
	_, ok = m.get(m.hasher(key), key)
	return
}

// Get `Maper` struct is used to retrieve the value associated with a given key
// from the hashmap. It takes a key as input and returns the corresponding value and a boolean
// indicating whether the key exists in the hashmap.
func (m *Maper[K, V]) Get(key K) (value V, ok bool) {
	return m.get(m.hasher(key), key)
}

func (m *Maper[K, V]) GetAndDelete(key K) (value V, ok bool) {
	var (
		h        = m.hasher(key)
		existing = m.metadata.Load().indexElement(h)
	)
	if existing == nil || existing.keyHash > h {
		existing = m.listHead.next()
	}
	for ; existing != nil && existing.keyHash <= h; existing = existing.next() {
		if existing.key == key {
			value, ok = *existing.value.Load(), !existing.isDeleted()
			if existing.remove() {
				m.removeItemFromIndex(existing)
			}
			return
		}
	}
	return
}

func (m *Maper[K, V]) get(h uintptr, key K) (value V, ok bool) {
	for elem := m.metadata.Load().indexElement(h); elem != nil && elem.keyHash <= h; elem = elem.nextPtr.Load() {
		if elem.key == key {
			ok = !elem.isDeleted()
			if ok {
				value = *elem.value.Load()
			}
			return
		}
	}
	return
}

// ProvideGet `Maper` struct is used to retrieve the value associated with a
// given key from the hashmap. If the key exists in the hashmap, the function returns the value and
// sets the `loaded` flag to true. If the key does not exist, the function calls the `provide` function
// to compute the value and sets the `computed` flag to true. The computed value is then added to the
// hashmap and returned.
func (m *Maper[K, V]) ProvideGet(key K, provide func() (V, bool)) (actual V, loaded, computed bool) {
	var (
		h        = m.hasher(key)
		data     = m.metadata.Load()
		existing = data.indexElement(h)
	)

	for elem := existing; elem != nil && elem.keyHash <= h; elem = elem.nextPtr.Load() {
		if elem.key == key {
			loaded = !elem.isDeleted()
			if loaded {
				actual = *elem.value.Load()
				return
			}
		}
	}

	var r bool
	_, _, _ = m.gsf.Do(strconv.FormatInt(int64(h), 10), func() (interface{}, error) {
		actual, loaded = provide()
		if loaded {
			m.Set(key, actual)
			computed = true
		}
		r = true
		return nil, nil
	})

	if !r {
		actual, loaded = m.get(h, key)
	}
	return
}

func (m *Maper[K, V]) GetOrSet(key K, value V) (actual V, loaded bool) {
	actual, loaded, _ = m.ProvideGet(key, func() (V, bool) {
		return value, true
	})
	return
}

func (m *Maper[K, V]) Set(key K, value V) {
	m.set(m.hasher(key), key, value)
}

func (m *Maper[K, V]) set(h uintptr, key K, value V) {
	var (
		created  bool
		valPtr   = &value
		alloc    *element[K, V]
		data     = m.metadata.Load()
		existing = data.indexElement(h)
	)

	if existing == nil || existing.keyHash > h {
		existing = m.listHead
	}

	if alloc, created = existing.inject(h, key, valPtr); alloc != nil {
		if created {
			m.numItems.Add(1)
		}
	} else {
		for existing = m.listHead; alloc == nil; alloc, created = existing.inject(h, key, valPtr) {
		}
		if created {
			m.numItems.Add(1)
		}
	}

	count := data.addItemToIndex(alloc)
	if resizeNeeded(uintptr(len(data.index)), count) && m.resizing.CAS(notResizing, resizingInProgress) {
		m.grow(0)
	}
}

// Swap `Maper` struct is used to atomically swap the value associated with a
// given key in the hashmap. It takes a key and a new value as input parameters and returns the old
// value that was swapped out and a boolean indicating whether the swap was successful.
func (m *Maper[K, V]) Swap(key K, newValue V) (oldValue V, swapped bool) {
	var (
		h        = m.hasher(key)
		existing = m.metadata.Load().indexElement(h)
	)

	if existing == nil || existing.keyHash > h {
		existing = m.listHead
	}

	if _, current, _ := existing.search(h, key); current != nil {
		oldValue, swapped = *current.value.Swap(&newValue), true
	} else {
		swapped = false
	}
	return
}

// CAS `Maper` struct is used to perform a Compare-and-Swap operation on a
// key-value pair in the hashmap.
func (m *Maper[K, V]) CAS(key K, oldValue, newValue V) bool {
	var (
		h        = m.hasher(key)
		existing = m.metadata.Load().indexElement(h)
	)

	if existing == nil || existing.keyHash > h {
		existing = m.listHead
	}

	if _, current, _ := existing.search(h, key); current != nil {
		if oldPtr := current.value.Load(); reflect.DeepEqual(*oldPtr, oldValue) {
			return current.value.CompareAndSwap(oldPtr, &newValue)
		}
	}
	return false
}

// ForEach `Maper` struct iterates over each key-value pair in the hashmap and
// applies a lambda function to each pair. The lambda function takes a key and value as input
// parameters and returns a boolean value. If the lambda function returns `true`, the iteration
// continues to the next key-value pair. If the lambda function returns `false`, the iteration stops.
func (m *Maper[K, V]) ForEach(lambda func(K, V) bool) {
	for item := m.listHead.next(); item != nil && lambda(item.key, *item.value.Load()); item = item.next() {
	}
}

func (m *Maper[K, V]) Grow(newSize uintptr) {
	if m.resizing.CAS(notResizing, resizingInProgress) {
		m.grow(newSize)
	}
}

func (m *Maper[K, V]) SetHasher(hasher func(K) uintptr) {
	m.hasher = hasher
}

func (m *Maper[K, V]) Len() uintptr {
	return m.numItems.Load()
}

func (m *Maper[K, V]) Clear() {
	index := make([]*element[K, V], m.defaultSize)
	header := (*reflect.SliceHeader)(unsafe.Pointer(&index))
	newdata := &metadata[K, V]{
		keyshifts: strconv.IntSize - log2(m.defaultSize),
		data:      unsafe.Pointer(header.Data),
		index:     index,
	}
	m.listHead.nextPtr.Store(nil)
	m.metadata.Store(newdata)
	m.numItems.Store(0)
}

// Keys returns the keys of the map.
func (m *Maper[K, V]) Keys() (keys []K) {
	keys = make([]K, m.Len())
	var (
		idx  = 0
		item = m.listHead.next()
	)
	for item != nil {
		keys[idx] = item.key
		idx++
		item = item.next()
	}
	return
}

// MarshalJSON convert the `Maper` object into a JSON-encoded byte slice.
func (m *Maper[K, V]) MarshalJSON() ([]byte, error) {
	gomap := make(map[K]V)
	for i := m.listHead.next(); i != nil; i = i.next() {
		gomap[i.key] = *i.value.Load()
	}
	return json.Marshal(gomap)
}

// UnmarshalJSON used to deserialize a JSON-encoded byte slice into a `Maper` object.
func (m *Maper[K, V]) UnmarshalJSON(i []byte) error {
	gomap := make(map[K]V)
	err := json.Unmarshal(i, &gomap)
	if err != nil {
		return err
	}
	for k, v := range gomap {
		m.Set(k, v)
	}
	return nil
}

// Fillrate calculates the fill rate of the hashmap.
// It returns the percentage of slots in the hashmap that are currently occupied by elements.
func (m *Maper[K, V]) Fillrate() uintptr {
	data := m.metadata.Load()

	return (data.count.Load() * 100) / uintptr(len(data.index))
}

func (m *Maper[K, V]) allocate(newSize uintptr) {
	if m.resizing.CAS(notResizing, resizingInProgress) {
		m.grow(newSize)
	}
}

func (m *Maper[K, V]) fillIndexItems(mapData *metadata[K, V]) {
	var (
		first     = m.listHead.next()
		item      = first
		lastIndex = uintptr(0)
	)

	for item != nil {
		index := item.keyHash >> mapData.keyshifts
		if item == first || index != lastIndex {
			mapData.addItemToIndex(item)
			lastIndex = index
		}
		item = item.next()
	}
}

func (m *Maper[K, V]) removeItemFromIndex(item *element[K, V]) {
	for {
		data := m.metadata.Load()
		index := item.keyHash >> data.keyshifts
		ptr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(data.data) + index*intSizeBytes))

		next := item.next()
		if next != nil && next.keyHash>>data.keyshifts != index {
			next = nil
		}

		swappedToNil := atomic.CompareAndSwapPointer(ptr, unsafe.Pointer(item), unsafe.Pointer(next)) && next == nil
		if data == m.metadata.Load() {
			m.numItems.Add(^uintptr(0))
			if swappedToNil {
				data.count.Add(^uintptr(0))
			}
			return
		}
	}
}

func (m *Maper[K, V]) grow(newSize uintptr) {
	for {
		currentStore := m.metadata.Load()
		if newSize == 0 {
			newSize = uintptr(len(currentStore.index)) << 1
		} else {
			newSize = roundUpPower2(newSize)
		}

		index := make([]*element[K, V], newSize)
		header := (*reflect.SliceHeader)(unsafe.Pointer(&index))

		newdata := &metadata[K, V]{
			keyshifts: strconv.IntSize - log2(newSize),
			data:      unsafe.Pointer(header.Data),
			count:     zutil.NewUintptr(0),
			index:     index,
		}

		m.fillIndexItems(newdata)
		m.metadata.Store(newdata)

		if !resizeNeeded(newSize, m.Len()) {
			m.resizing.Store(notResizing)
			return
		}
		newSize = 0
	}
}

func (md *metadata[K, V]) indexElement(hashedKey uintptr) *element[K, V] {
	index := hashedKey >> md.keyshifts
	ptr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(md.data) + index*intSizeBytes))
	item := (*element[K, V])(atomic.LoadPointer(ptr))
	for (item == nil || hashedKey < item.keyHash || item.isDeleted()) && index > 0 {
		index--
		ptr = (*unsafe.Pointer)(unsafe.Pointer(uintptr(md.data) + index*intSizeBytes))
		item = (*element[K, V])(atomic.LoadPointer(ptr))
	}
	return item
}

func (md *metadata[K, V]) addItemToIndex(item *element[K, V]) uintptr {
	index := item.keyHash >> md.keyshifts
	ptr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(md.data) + index*intSizeBytes))
	for {
		elem := (*element[K, V])(atomic.LoadPointer(ptr))
		if elem == nil {
			if atomic.CompareAndSwapPointer(ptr, nil, unsafe.Pointer(item)) {
				return md.count.Add(1)
			}
			continue
		}
		if item.keyHash < elem.keyHash {
			if !atomic.CompareAndSwapPointer(ptr, unsafe.Pointer(elem), unsafe.Pointer(item)) {
				continue
			}
		}
		return 0
	}
}

func resizeNeeded(length, count uintptr) bool {
	return (count*100)/length > maxFillRate
}

func roundUpPower2(i uintptr) uintptr {
	i--
	i |= i >> 1
	i |= i >> 2
	i |= i >> 4
	i |= i >> 8
	i |= i >> 16
	i |= i >> 32
	i++
	return i
}

func log2(i uintptr) (n uintptr) {
	for p := uintptr(1); p < i; p, n = p<<1, n+1 {
	}
	return
}
