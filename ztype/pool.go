package ztype

import (
	"sync"
)

// mapSlicePool []map[string]interface{} slice pool for ToMaps
var mapSlicePool = sync.Pool{
	New: func() interface{} {
		return make([]map[string]interface{}, 0, 4)
	},
}

// stringSlicePool string slice pool for tag option parsing
var stringSlicePool = sync.Pool{
	New: func() interface{} {
		return make([]string, 0, 8)
	},
}

// interfaceSlicePool interface{} slice pool for type conversions
var interfaceSlicePool = sync.Pool{
	New: func() interface{} {
		return make([]interface{}, 0, 8)
	},
}

// intSlicePool int slice pool for numeric conversions
var intSlicePool = sync.Pool{
	New: func() interface{} {
		return make([]int, 0, 8)
	},
}

// getMapSlice gets a map slice from object pool
func getMapSlice() []map[string]interface{} {
	s := mapSlicePool.Get().([]map[string]interface{})
	return s[:0]
}

// putMapSlice returns a map slice to object pool
func putMapSlice(s []map[string]interface{}) {
	if s == nil || cap(s) > 64 {
		return
	}

	for i := range s {
		s[i] = nil
	}
	s = s[:0]
	mapSlicePool.Put(s)
}

// getStringSlice gets a string slice from object pool
func getStringSlice() []string {
	s := stringSlicePool.Get().([]string)
	return s[:0]
}

// putStringSlice returns a string slice to object pool
func putStringSlice(s []string) {
	if s == nil || cap(s) > 64 {
		return
	}
	for i := range s {
		s[i] = ""
	}
	s = s[:0]
	stringSlicePool.Put(s)
}

// getInterfaceSlice gets an interface{} slice from object pool
func getInterfaceSlice() []interface{} {
	s := interfaceSlicePool.Get().([]interface{})
	return s[:0]
}

// putInterfaceSlice returns an interface{} slice to object pool
func putInterfaceSlice(s []interface{}) {
	if s == nil || cap(s) > 64 {
		return
	}
	for i := range s {
		s[i] = nil
	}
	s = s[:0]
	interfaceSlicePool.Put(s)
}

// getIntSlice gets an int slice from object pool
func getIntSlice() []int {
	s := intSlicePool.Get().([]int)
	return s[:0]
}

// putIntSlice returns an int slice to object pool
func putIntSlice(s []int) {
	if s == nil || cap(s) > 64 {
		return
	}
	for i := range s {
		s[i] = 0
	}
	s = s[:0]
	intSlicePool.Put(s)
}
