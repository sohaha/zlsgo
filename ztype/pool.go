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
		return make([]string, 0, 4)
	},
}

// getMapSlice gets a map slice from object pool
func getMapSlice() []map[string]interface{} {
	s := mapSlicePool.Get().([]map[string]interface{})
	return s[:0]
}

// putMapSlice returns a map slice to object pool
func putMapSlice(s []map[string]interface{}) {
	if s == nil {
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
	if s == nil {
		return
	}
	for i := range s {
		s[i] = ""
	}
	s = s[:0]
	stringSlicePool.Put(s)
}

// resetMap resets map for reuse
func resetMap(m map[string]interface{}) {
	for k := range m {
		delete(m, k)
	}
}
