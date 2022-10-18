//go:build go1.18
// +build go1.18

package zarray

import (
	"errors"
	"reflect"
	"sort"
	"strconv"
	"sync/atomic"
	"unsafe"

	"github.com/sohaha/zlsgo/zutil"
	"golang.org/x/sync/singleflight"
)

const (
	// defaultSize is the default size for a zero allocated map
	defaultSize = 256

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
		int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | uintptr | float32 | float64 | string | complex64 | complex128
	}

	metadata[K hashable, V any] struct {
		count     *zutil.Uintptr
		data      unsafe.Pointer
		index     []*element[K, V]
		keyshifts uintptr
	}

	// Map implements the concurrent hashmap
	StMap[K hashable, V any] struct {
		g        singleflight.Group
		listHead *element[K, V]
		hasher   func(K) uintptr
		metadata atomicPointer[metadata[K, V]]
		resizing *zutil.Uint32
		numItems *zutil.Uintptr
	}

	deletionRequest[K hashable] struct {
		key     K
		keyHash uintptr
	}
)

func NewHashMap[K hashable, V any](size ...uintptr) *StMap[K, V] {
	m := &StMap[K, V]{
		listHead: newListHead[K, V](),
		resizing: zutil.NewUint32(0),
		numItems: zutil.NewUintptr(0),
	}
	m.numItems.Store(0)
	if len(size) > 0 {
		m.allocate(size[0])
	} else {
		m.allocate(defaultSize)
	}
	m.setDefaultHasher()
	return m
}

func (m *StMap[K, V]) Delele(keys ...K) {
	if len(keys) == 0 {
		return
	}
	var (
		size = len(keys)
		delQ = make([]deletionRequest[K], size)
		elem = m.listHead.next()
		iter = 0
	)
	for idx := 0; idx < size; idx++ {
		delQ[idx].keyHash, delQ[idx].key = m.hasher(keys[idx]), keys[idx]
	}

	sort.Slice(delQ, func(i, j int) bool {
		return delQ[i].keyHash < delQ[j].keyHash
	})

	for elem != nil && iter < size {
		if elem.keyHash == delQ[iter].keyHash && elem.key == delQ[iter].key {
			elem.remove()
			m.removeItemFromIndex(elem)
			iter++
			elem = elem.next()
		} else if elem.keyHash > delQ[iter].keyHash {
			iter++
		} else {
			elem = elem.next()
		}
	}

	for iter := m.listHead; iter != nil; iter = iter.next() {
	}
}

func (m *StMap[K, V]) get(key K) (h uintptr, value V, ok bool) {
	h = m.hasher(key)
	for elem := m.metadata.Load().indexElement(h); elem != nil; elem = elem.nextPtr.Load() {
		if elem.keyHash == h && elem.key == key {
			value, ok = *elem.value.Load(), true
			return
		}
		if elem.keyHash <= h || elem.keyHash == marked {
			continue
		} else {
			break
		}
	}
	ok = false
	return
}

func (m *StMap[K, V]) Get(key K) (value V, ok bool) {
	_, value, ok = m.get(key)
	return
}

func (m *StMap[K, V]) ProvideGet(key K, provide func(K) (V, bool)) (value V, ok bool) {
	var h uintptr
	h, value, ok = m.get(key)
	if !ok {
		v, err, _ := m.g.Do(strconv.Itoa(int(h)), func() (interface{}, error) {
			value, ok := provide(key)
			if !ok {
				return nil, errors.New("key not found")
			}

			m.Set(key, value)
			return value, nil
		})

		if err == nil {
			ok = true
			value = v.(V)
		}
	}
	return
}

func (m *StMap[K, V]) Set(key K, value V) {
	var (
		created  bool
		h        = m.hasher(key)
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

func (m *StMap[K, V]) ForEach(lambda func(K, V) bool) {
	for item := m.listHead.next(); item != nil && lambda(item.key, *item.value.Load()); item = item.next() {
	}
}

func (m *StMap[K, V]) Grow(newSize uintptr) {
	if m.resizing.CAS(notResizing, resizingInProgress) {
		m.grow(newSize)
	}
}

func (m *StMap[K, V]) SetHasher(hs func(K) uintptr) {
	m.hasher = hs
}

func (m *StMap[K, V]) Len() uintptr {
	return m.numItems.Load()
}

func (m *StMap[K, V]) Fillrate() uintptr {
	data := m.metadata.Load()
	return (data.count.Load() * 100) / uintptr(len(data.index))
}

func (m *StMap[K, V]) allocate(newSize uintptr) {
	if m.resizing.CAS(notResizing, resizingInProgress) {
		m.grow(newSize)
	}
}

func (m *StMap[K, V]) fillIndexItems(mapData *metadata[K, V]) {
	first := m.listHead
	item := first
	lastIndex := uintptr(0)
	for item != nil {
		index := item.keyHash >> mapData.keyshifts
		if item == first || index != lastIndex {
			mapData.addItemToIndex(item)
			lastIndex = index
		}
		item = item.next()
	}
}

func (m *StMap[K, V]) removeItemFromIndex(item *element[K, V]) {
	for {
		data := m.metadata.Load()
		index := item.keyHash >> data.keyshifts
		ptr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(data.data) + index*intSizeBytes))

		next := item.next()
		if next != nil && item.keyHash>>data.keyshifts != index {
			next = nil
		}
		atomic.CompareAndSwapPointer(ptr, unsafe.Pointer(item), unsafe.Pointer(next))

		if data == m.metadata.Load() {
			m.numItems.Add(marked)
			return
		}
	}
}

func (m *StMap[K, V]) grow(newSize uintptr) {
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

		if !resizeNeeded(newSize, uintptr(m.Len())) {
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
	for (item == nil || hashedKey < item.keyHash) && index > 0 {
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
