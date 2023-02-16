package zcache

import (
	"sync"
	"time"
	"unsafe"

	"github.com/sohaha/zlsgo/ztime"
	"golang.org/x/sync/singleflight"
)

// FastCache concurrent LRU cache structure
type FastCache struct {
	gsf        singleflight.Group
	callback   handler
	locks      []sync.Mutex
	insts      [][2]*lruCache
	expiration time.Duration
	mask       int32
}

type Option struct {
	Callback   func(ActionKind, string, uintptr)
	Expiration time.Duration
	Bucket     uint16
	Cap        uint16
	LRU2Cap    uint16
}

// NewFast Fast LRU cache
func NewFast(opt ...func(o *Option)) *FastCache {
	o := Option{
		Cap:    1 << 10,
		Bucket: 4,
	}

	for _, f := range opt {
		f(&o)
	}

	var mask uint16
	if o.Bucket > 0 && o.Bucket&(o.Bucket-1) == 0 {
		mask = o.Bucket - 1
	} else {
		o.Bucket |= o.Bucket >> 1
		o.Bucket |= o.Bucket >> 2
		o.Bucket |= o.Bucket >> 4
		mask = o.Bucket | (o.Bucket >> 8)
	}

	c := &FastCache{
		locks:    make([]sync.Mutex, mask+1),
		insts:    make([][2]*lruCache, mask+1),
		callback: o.Callback,
		mask:     int32(mask),
	}

	for i := range c.insts {
		c.insts[i][0] = &lruCache{dlList: make([][2]uint16, uint32(o.Cap)+1), nodes: make([]node, o.Cap), hashmap: make(map[string]uint16, o.Cap), last: 0}
		if o.LRU2Cap > 0 {
			c.insts[i][1] = &lruCache{dlList: make([][2]uint16, uint32(o.LRU2Cap)+1), nodes: make([]node, o.LRU2Cap), hashmap: make(map[string]uint16, o.LRU2Cap), last: 0}
		}
	}

	if o.Expiration > 0 {
		c.expiration = o.Expiration
	}
	return c
}

func (l *FastCache) set(k string, v *interface{}, b []byte, expiration ...time.Duration) {
	if l.callback != nil {
		if v != nil {
			l.callback(SET, k, uintptr(unsafe.Pointer(v)))
		} else {
			l.callback(SET, k, uintptr(unsafe.Pointer(&b)))
		}
	}
	idx := hasher(k) & l.mask
	var expireAt int64
	if len(expiration) > 0 && expiration[0] > 0 {
		expireAt = ztime.Clock()*1000 + int64(expiration[0])
	} else if l.expiration > 0 {
		expireAt = ztime.Clock()*1000 + int64(l.expiration)
	}
	l.locks[idx].Lock()
	l.insts[idx][0].put(k, v, b, expireAt)
	l.locks[idx].Unlock()
}

// Set an item into cache
func (l *FastCache) Set(key string, val interface{}, expiration ...time.Duration) {
	l.set(key, &val, nil, expiration...)
}

// SetBytes an item into cache
func (l *FastCache) SetBytes(key string, b []byte) {
	l.set(key, nil, b)
}

// Get value of key from cache with result
func (l *FastCache) Get(key string) (interface{}, bool) {
	if i, b, ok := l.get(key); ok {
		if i != nil {
			return *i, true
		}
		return b, true
	}
	return nil, false
}

// GetBytes value of key from cache with result
func (l *FastCache) GetBytes(key string) ([]byte, bool) {
	if i, b, ok := l.get(key); ok {
		if b != nil {
			return b, true
		}
		b, ok = (*i).([]byte)
		return b, ok
	}
	return nil, false
}

// ProvideGet get value of key from cache with result and provide default value
func (l *FastCache) ProvideGet(key string, provide func() (interface{}, bool)) (interface{}, bool) {
	if i, _, ok := l.get(key); ok && i != nil {
		return *i, true
	}

	_, _, _ = l.gsf.Do(key, func() (value interface{}, err error) {
		value, ok := provide()
		if ok {
			l.Set(key, value)
		}
		return
	})

	return l.Get(key)
}

func (l *FastCache) getValue(key string, idx, level int32) (*node, int) {
	n, s := l.insts[idx][level].get(key)
	if s > 0 && !n.isDelete && (n.expireAt == 0 || (ztime.Clock()*1000 <= n.expireAt)) {
		return n, s
	}
	return nil, 0
}

func (l *FastCache) get(key string) (i *interface{}, b []byte, loaded bool) {
	idx := hasher(key) & l.mask
	l.locks[idx].Lock()
	n, s := (*node)(nil), 0
	if l.insts[idx][1] == nil {
		n, s = l.getValue(key, idx, 0)
	} else {
		e := int64(0)
		if n, s, e = l.insts[idx][0].delete(key); s <= 0 {
			n, s = l.getValue(key, idx, 1)
		} else {
			l.insts[idx][1].put(key, n.value.value, n.value.byteValue, e)
		}
	}
	if s <= 0 {
		l.locks[idx].Unlock()
		if l.callback != nil {
			l.callback(GET, key, uintptr(0))
		}
		return
	}
	i, b = n.value.value, n.value.byteValue
	l.locks[idx].Unlock()
	if l.callback != nil {
		if i != nil {
			l.callback(GET, key, uintptr(unsafe.Pointer(i)))
		} else {
			var b interface{} = b
			l.callback(GET, key, uintptr(unsafe.Pointer(&b)))
		}
	}
	return i, b, true
}

// Delete item by key from cache
func (l *FastCache) Delete(key string) {
	idx := hasher(key) & l.mask
	l.locks[idx].Lock()
	n, s, e := l.insts[idx][0].delete(key)
	if l.insts[idx][1] != nil {
		if n2, s2, e2 := l.insts[idx][1].delete(key); n2 != nil && (n == nil || e < e2) {
			n, s = n2, s2
		}
	}
	if s > 0 {
		if l.callback != nil {
			if n.value.value != nil {
				l.callback(DELETE, key, uintptr(unsafe.Pointer(n.value.value)))
			} else {
				l.callback(DELETE, key, uintptr(unsafe.Pointer(&n.value.byteValue)))
			}
		}
		n.value.value, n.value.byteValue = nil, nil
	} else if l.callback != nil {
		l.callback(DELETE, key, uintptr(0))
	}
	l.locks[idx].Unlock()
}

// ForEach walk through all items in cache
func (l *FastCache) ForEach(walker func(key string, iface interface{}) bool) {
	for i := range l.insts {
		l.locks[i].Lock()
		if l.insts[i][0].forEach(walker); l.insts[i][1] != nil {
			l.insts[i][1].forEach(walker)
		}
		l.locks[i].Unlock()
	}
}
