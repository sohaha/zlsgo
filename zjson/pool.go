package zjson

import (
	"sync"
)

// pathCachePool reuses path parsing results
var pathCachePool = sync.Pool{
	New: func() interface{} {
		return make([]pathResult, 0, 8)
	},
}

func getPathCache() []pathResult {
	return pathCachePool.Get().([]pathResult)[:0]
}

func putPathCache(cache []pathResult) {
	if cap(cache) > 32 {
		return
	}
	pathCachePool.Put(cache)
}
