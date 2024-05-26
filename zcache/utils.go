package zcache

import (
	"github.com/sohaha/zlsgo/ztime"
)

type ActionKind int

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

const (
	SET ActionKind = iota + 1
	GET
	DELETE
)

type handler func(action ActionKind, key string, valuePtr uintptr)

// Callback set callback function for cache
func (l *FastCache) Callback(h handler) {
	old := l.callback
	l.callback = func(action ActionKind, key string, valuePtr uintptr) {
		old(action, key, valuePtr)
		h(action, key, valuePtr)
	}
}

var p, n = uint16(0), uint16(1)

type (
	value struct {
		value     *interface{}
		byteValue []byte
	}

	node struct {
		key      string
		value    value
		expireAt int64
		isDelete bool
	}

	lruCache struct {
		hashmap map[string]uint16
		dlList  [][2]uint16
		nodes   []node
		last    uint16
	}
)

func (c *lruCache) put(k string, i *interface{}, b []byte, expireAt int64) int {
	if x, ok := c.hashmap[k]; ok {
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
	return 1
}

func (c *lruCache) get(k string) (*node, int) {
	if x, ok := c.hashmap[k]; ok {
		c.adjust(x, p, n)
		return &c.nodes[x-1], 1
	}
	return nil, 0
}

func (c *lruCache) delete(k string) (_ *node, _ int, e int64) {
	if x, ok := c.hashmap[k]; ok && !c.nodes[x-1].isDelete {
		c.nodes[x-1].expireAt, c.nodes[x-1].isDelete, e = 0, true, c.nodes[x-1].expireAt
		c.adjust(x, n, p)
		return &c.nodes[x-1], 1, e
	}
	return nil, 0, 0
}

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

func (c *lruCache) adjust(idx, f, t uint16) {
	if c.dlList[idx][f] != 0 {
		c.dlList[c.dlList[idx][t]][f], c.dlList[c.dlList[idx][f]][t], c.dlList[idx][f], c.dlList[idx][t], c.dlList[c.dlList[0][t]][f], c.dlList[0][t] = c.dlList[idx][f], c.dlList[idx][t], 0, c.dlList[0][t], idx, idx
	}
}

func hasher(s string) (hash int32) {
	for i := 0; i < len(s); i++ {
		hash = hash*131 + int32(s[i])
	}
	return hash
}
