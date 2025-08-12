package fast

import (
	"github.com/sohaha/zlsgo/ztime"
)

// ActionKind represents the type of operation performed on a cache item
type ActionKind int

// String returns the string representation of an ActionKind
func (k ActionKind) String() string {
	switch k {
	case GET:
		return "GET"
	case SET:
		return "SET"
	case DELETE:
		return "DELETE"
	default:
		return "UNKNOWN"
	}
}

// Cache operation constants
const (
	// SET represents a cache item creation or update operation
	SET ActionKind = iota + 1
	// GET represents a cache item retrieval operation
	GET
	// DELETE represents a cache item removal operation
	DELETE
)

// handler is a callback function type for cache operations
// action: the type of operation performed
// key: the cache key involved in the operation
// valuePtr: pointer to the value involved in the operation
type handler func(action ActionKind, key string, valuePtr uintptr)

// Callback sets a handler function to be called for each cache operation.
// The new handler is chained with any existing handler, so both will be executed.
func (l *FastCache) Callback(h handler) {
	old := l.callback
	l.callback = func(action ActionKind, key string, valuePtr uintptr) {
		old(action, key, valuePtr)
		h(action, key, valuePtr)
	}
}

// p and n are indices for the previous and next pointers in the doubly linked list
var p, n = uint16(0), uint16(1)

type (
	// value stores either an interface{} pointer or a byte slice
	value struct {
		value     *interface{}
		byteValue []byte
	}

	// node represents a cache entry in the LRU cache
	node struct {
		key      string
		value    value
		expireAt int64
		isDelete bool
	}

	// lruCache implements a Least Recently Used cache with a fixed capacity
	lruCache struct {
		hashmap map[string]uint16
		dlList  [][2]uint16
		nodes   []node
		last    uint16
		size    int
	}
)

// put adds or updates an item in the LRU cache.
// Returns 0 if the item was updated, 1 if it was added.
func (c *lruCache) put(k string, i *interface{}, b []byte, expireAt int64) int {
	if x, ok := c.hashmap[k]; ok {
		if c.nodes[x-1].isDelete {
			c.size++
		}
		c.nodes[x-1].value.value, c.nodes[x-1].value.byteValue, c.nodes[x-1].expireAt, c.nodes[x-1].isDelete = i, b, expireAt, false
		c.adjust(x, p, n)
		return 0
	}

	if c.last == uint16(cap(c.nodes)) {
		tail := &c.nodes[c.dlList[0][p]-1]
		delete(c.hashmap, (*tail).key)
		c.hashmap[k], (*tail).key, (*tail).value.value, (*tail).value.byteValue, (*tail).expireAt, (*tail).isDelete = c.dlList[0][p], k, i, b, expireAt, false
		c.adjust(c.dlList[0][p], p, n)
		return 1
	}

	c.last++
	if len(c.hashmap) <= 0 {
		c.dlList[0][p] = c.last
	} else {
		c.dlList[c.dlList[0][n]][p] = c.last
	}
	c.nodes[c.last-1].key, c.nodes[c.last-1].value.value, c.nodes[c.last-1].value.byteValue, c.nodes[c.last-1].expireAt, c.nodes[c.last-1].isDelete, c.dlList[c.last], c.hashmap[k], c.dlList[0][n] = k, i, b, expireAt, false, [2]uint16{0, c.dlList[0][n]}, c.last, c.last
	c.size++
	return 1
}

// get retrieves an item from the LRU cache by its key.
// Returns the node and 1 if found, nil and 0 otherwise.
func (c *lruCache) get(k string) (*node, int) {
	if x, ok := c.hashmap[k]; ok {
		c.adjust(x, p, n)
		return &c.nodes[x-1], 1
	}
	return nil, 0
}

// delete removes an item from the LRU cache by its key.
// Returns the node, 1, and the expiration time if found and not already deleted,
// otherwise returns nil, 0, and 0.
func (c *lruCache) delete(k string) (_ *node, _ int, e int64) {
	if x, ok := c.hashmap[k]; ok && !c.nodes[x-1].isDelete {
		c.nodes[x-1].expireAt, c.nodes[x-1].isDelete, e = 0, true, c.nodes[x-1].expireAt
		c.adjust(x, n, p)
		if c.size > 0 {
			c.size--
		}
		return &c.nodes[x-1], 1, e
	}
	return nil, 0, 0
}

// forEach iterates through all non-deleted and non-expired items in the LRU cache
// and applies the provided function to each key-value pair.
// The iteration continues as long as the function returns true, and stops when it returns false.
func (c *lruCache) forEach(walker func(key string, iface interface{}) bool) {
	for idx := c.dlList[0][n]; idx != 0; idx = c.dlList[idx][n] {
		if !c.nodes[idx-1].isDelete {
			if c.nodes[idx-1].expireAt != 0 && (ztime.Clock()*1000 >= c.nodes[idx-1].expireAt) {
				n, _, _ := c.delete(c.nodes[idx-1].key)
				n.value.value, n.value.byteValue = nil, nil
				continue
			}
			if c.nodes[idx-1].value.byteValue != nil {
				if !walker(c.nodes[idx-1].key, c.nodes[idx-1].value.byteValue) {
					return
				}
			} else {
				if !walker(c.nodes[idx-1].key, *c.nodes[idx-1].value.value) {
					return
				}
			}
		}
	}
}

// cleanExpired removes expired items without invoking a walker.
// It traverses the list and deletes nodes whose expireAt <= now.
func (c *lruCache) cleanExpired(now int64) {
	for idx := c.dlList[0][n]; idx != 0; idx = c.dlList[idx][n] {
		if c.nodes[idx-1].isDelete {
			continue
		}
		if c.nodes[idx-1].expireAt != 0 && now >= c.nodes[idx-1].expireAt {
			n, _, _ := c.delete(c.nodes[idx-1].key)
			n.value.value, n.value.byteValue = nil, nil
		}
	}
}

// isEmpty reports whether the cache currently holds any non-deleted items.
func (c *lruCache) isEmpty() bool {
	return c.size == 0
}

// adjust reorders the doubly linked list to move the specified node
// to the most recently used position.
func (c *lruCache) adjust(idx, f, t uint16) {
	if c.dlList[idx][f] != 0 {
		c.dlList[c.dlList[idx][t]][f], c.dlList[c.dlList[idx][f]][t], c.dlList[idx][f], c.dlList[idx][t], c.dlList[c.dlList[0][t]][f], c.dlList[0][t] = c.dlList[idx][f], c.dlList[idx][t], 0, c.dlList[0][t], idx, idx
	}
}

// hasher computes a 32-bit hash value for a string using a simple algorithm.
// This is used to determine the bucket index for cache sharding.
func hasher(s string) (hash uint16) {
	for i := 0; i < len(s); i++ {
		hash = hash*131 + uint16(s[i])
	}
	return hash
}
