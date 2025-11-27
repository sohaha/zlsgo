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
	// hashable defines the types that can be used as keys in the hashmap
	hashable interface {
		constraints.Integer | constraints.Float | constraints.Complex | ~string | uintptr | unsafe.Pointer
	}

	// metadata contains internal data structures for the hashmap implementation
	metadata[K hashable, V any] struct {
		count     *zutil.Uintptr
		data      unsafe.Pointer
		index     []*element[K, V]
		keyshifts uintptr
	}

	// Maper implements a concurrent hashmap with type-safe generic key-value pairs
	// It provides thread-safe operations for storing, retrieving, and manipulating data
	Maper[K hashable, V any] struct {
		gsf         singleflight.Group
		listHead    *element[K, V]
		hasher      func(K) uintptr
		metadata    atomicPointer[metadata[K, V]]
		resizing    *zutil.Uint32
		numItems    *zutil.Uintptr
		defaultSize uintptr
	}

	// deletionRequest represents a key scheduled for deletion from the hashmap
	deletionRequest[K hashable] struct {
		key     K
		keyHash uintptr
	}
)

type provideResult[V any] struct {
	value V
	ok    bool
}

func makeSingleflightKey[K hashable](hash uintptr, key K) string {
	keyStr := zutil.KeySignature(key)
	return "h" + strconv.FormatUint(uint64(hash), 16) + ":" + keyStr
}

// NewHashMap creates a new concurrent hashmap with the specified initial size.
// If no size is provided, a default size is used.
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

// Delete removes one or more key-value pairs from the map.
// If multiple keys are provided, they are processed in an optimized batch operation.
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

// Has checks if a key exists in the map.
// Returns true if the key exists, false otherwise.
func (m *Maper[K, V]) Has(key K) (ok bool) {
	_, ok = m.get(m.hasher(key), key)
	return
}

// Get retrieves the value associated with the specified key.
// Returns the value and a boolean indicating whether the key was found in the map.
func (m *Maper[K, V]) Get(key K) (value V, ok bool) {
	return m.get(m.hasher(key), key)
}

// GetAndDelete retrieves the value associated with the specified key and removes it from the map.
// Returns the value and a boolean indicating whether the key was found and removed.
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

// If the key exists, returns the value with loaded=true.
// If the key doesn't exist, calls the provide function to compute a value,
// stores it in the map if the provider returns true, and returns with computed=true.
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

	keyStr := makeSingleflightKey(h, key)
	v, err, shared := m.gsf.Do(keyStr, func() (interface{}, error) {
		value, ok := provide()
		if ok {
			m.Set(key, value)
		}
		return provideResult[V]{value: value, ok: ok}, nil
	})
	if err != nil {
		return
	}

	if res, ok := v.(provideResult[V]); ok {
		if res.ok {
			if !shared {
				actual = res.value
				loaded = true
				computed = true
				return
			}

			actual, loaded = m.get(h, key)
			computed = true
			return
		}
	}

	actual, loaded = m.get(h, key)
	return
}

// GetOrSet retrieves the existing value for a key if present, otherwise stores and returns the given value.
// Returns the actual value stored and a boolean indicating whether the value was loaded (true) or stored (false).
func (m *Maper[K, V]) GetOrSet(key K, value V) (actual V, loaded bool) {
	actual, loaded, _ = m.ProvideGet(key, func() (V, bool) {
		return value, true
	})
	return
}

// Set adds or updates a key-value pair in the map.
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

// Swap atomically replaces the value for a key and returns the previous value.
// Returns the old value and a boolean indicating whether the swap was successful.
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

// CAS (Compare-And-Swap) atomically replaces the value for a key if it matches the old value.
// Returns true if the swap was performed, false otherwise.
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

// ForEach iterates over all key-value pairs in the map and applies the provided function to each.
// The iteration continues as long as the lambda function returns true, and stops when it returns false.
func (m *Maper[K, V]) ForEach(lambda func(K, V) bool) {
	for item := m.listHead.next(); item != nil && lambda(item.key, *item.value.Load()); item = item.next() {
	}
}

// Grow increases the size of the map to accommodate more elements efficiently.
// This operation is performed concurrently with other map operations.
func (m *Maper[K, V]) Grow(newSize uintptr) {
	if m.resizing.CAS(notResizing, resizingInProgress) {
		m.grow(newSize)
	}
}

// SetHasher sets a custom hash function for the map's keys.
// This should be called before any other operations on the map.
func (m *Maper[K, V]) SetHasher(hasher func(K) uintptr) {
	m.hasher = hasher
}

// Len returns the number of key-value pairs in the map.
func (m *Maper[K, V]) Len() uintptr {
	return m.numItems.Load()
}

// Clear removes all key-value pairs from the map, resetting it to an empty state.
func (m *Maper[K, V]) Clear() {
	index := make([]*element[K, V], m.defaultSize)
	header := (*reflect.SliceHeader)(unsafe.Pointer(&index))
	newdata := &metadata[K, V]{
		keyshifts: strconv.IntSize - log2(m.defaultSize),
		data:      unsafe.Pointer(header.Data),
		index:     index,
		count:     zutil.NewUintptr(0),
	}
	m.listHead.nextPtr.Store(nil)
	m.metadata.Store(newdata)
	m.numItems.Store(0)
}

// Keys returns a slice containing all keys currently in the map.
// The order of keys in the returned slice is not guaranteed.
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

// Values returns a slice containing all values currently in the map.
// The order of values in the returned slice is not guaranteed.
func (m *Maper[K, V]) Values() (values []V) {
	values = make([]V, m.Len())
	var (
		idx  = 0
		item = m.listHead.next()
	)
	for item != nil {
		values[idx] = *item.value.Load()
		idx++
		item = item.next()
	}
	return
}

// MarshalJSON implements the json.Marshaler interface to convert the map into a JSON-encoded byte slice.
// This allows the map to be serialized to JSON format.
func (m *Maper[K, V]) MarshalJSON() ([]byte, error) {
	gomap := make(map[K]V)
	for i := m.listHead.next(); i != nil; i = i.next() {
		gomap[i.key] = *i.value.Load()
	}
	return json.Marshal(gomap)
}

// UnmarshalJSON implements the json.Unmarshaler interface to deserialize a JSON-encoded byte slice
// into the map. This allows the map to be reconstructed from JSON format.
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

// Fillrate calculates and returns the current fill rate of the map as a percentage.
// This indicates how full the underlying data structure is relative to its capacity.
func (m *Maper[K, V]) Fillrate() uintptr {
	data := m.metadata.Load()

	return (data.count.Load() * 100) / uintptr(len(data.index))
}

// allocate initializes the map's internal data structures with the specified size.
func (m *Maper[K, V]) allocate(newSize uintptr) {
	if m.resizing.CAS(notResizing, resizingInProgress) {
		m.grow(newSize)
	}
}

// fillIndexItems populates the index with all current elements in the map.
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

// removeItemFromIndex removes an element from the map's index.
// This is called when an element is deleted from the map.
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

// grow resizes the map's internal data structures to the specified size.
// This is called when the map needs to be expanded to accommodate more elements.
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

// indexElement finds the element in the index that corresponds to the given hash key.
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

// addItemToIndex adds an element to the map's index and returns its position.
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

// resizeNeeded determines if the map needs to be resized based on its current fill rate.
func resizeNeeded(length, count uintptr) bool {
	return (count*100)/length > maxFillRate
}

// roundUpPower2 rounds up a number to the next power of 2.
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

// log2 calculates the base-2 logarithm of a number.
func log2(i uintptr) (n uintptr) {
	for p := uintptr(1); p < i; p, n = p<<1, n+1 {
	}
	return
}
