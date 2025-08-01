package zcache

import (
	"time"

	"github.com/sohaha/zlsgo/zcache/fast"
	"github.com/sohaha/zlsgo/ztype"
)

// FastCache implements a high-performance concurrent LRU (Least Recently
// Used) cache with support for expiration, callbacks, and multiple buckets
// for reduced lock contention.
type FastCache struct {
	c *fast.FastCache
}

// Options defines configuration parameters for creating a new FastCache instance.
type Options fast.Options

// NewFast creates a new FastCache instance with the specified options.
// If no options are provided, default values are used.
func NewFast(opt ...func(o *Options)) *FastCache {
	o := Options{
		Cap:    1 << 10,
		Bucket: 4,
	}

	for _, f := range opt {
		f(&o)
	}

	c := fast.NewFast(func(fo *fast.Options) {
		*fo = fast.Options(o)
	})

	return &FastCache{
		c: c,
	}
}

// Set adds or updates an item in the cache with the specified key, value,
// and optional expiration. If no expiration is provided, the default
// expiration time is used (if configured).
func (l *FastCache) Set(key string, val interface{}, expiration ...time.Duration) {
	l.c.Set(key, val, expiration...)
}

// SetBytes adds or updates a byte slice in the cache with the specified key.
// The default expiration time is used (if configured).
func (l *FastCache) SetBytes(key string, b []byte) {
	l.c.SetBytes(key, b)
}

// Get retrieves an item from the cache by its key. It returns the item's
// value and a boolean indicating whether the item was found.
func (l *FastCache) Get(key string) (interface{}, bool) {
	return l.c.Get(key)
}

// GetAny retrieves an item from the cache and wraps it in a ztype.Type for
// flexible type conversion. It returns the wrapped value and a boolean
// indicating whether the item was found.
func (l *FastCache) GetAny(key string) (ztype.Type, bool) {
	if v, ok := l.Get(key); ok {
		return ztype.New(v), true
	}
	return ztype.Type{}, false
}

// GetBytes retrieves a byte slice from the cache by its key. It returns the
// byte slice and a boolean indicating whether the item was found and is a
// byte slice.
func (l *FastCache) GetBytes(key string) ([]byte, bool) {
	return l.c.GetBytes(key)
}

// ProvideGet retrieves an item from the cache, or computes and stores it if
// not present. If the item doesn't exist, the provide function is called to
// generate the value. It returns the item's value and a boolean indicating
// whether the item was found or created.
func (l *FastCache) ProvideGet(key string, provide func() (interface{}, bool), expiration ...time.Duration) (interface{}, bool) {
	return l.c.ProvideGet(key, provide, expiration...)
}

// Delete removes an item with the specified key from the cache. If the item
// doesn't exist, this operation is a no-op.
func (l *FastCache) Delete(key string) {
	l.c.Delete(key)
}

// ForEach iterates through all items in the cache and applies the provided
// function to each key-value pair. The iteration continues as long as the
// function returns true, and stops when it returns false.
func (l *FastCache) ForEach(walker func(key string, iface interface{}) bool) {
	l.c.ForEach(walker)
}
