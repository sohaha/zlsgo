package zjson

import (
	"sync"
)

// maxPathCacheCap bounds the capacity kept in the pool so an unusually deep
// path does not pin a large backing array.
const maxPathCacheCap = 32

// pathCachePool reuses path parsing results. The pool stores a *[]pathResult
// and callers own that pointer between get and put, appending through it
// (append(*p, ...)) so the slice header - and any capacity grown by a deep
// path - stays with the pooled entry. Handing the pointer back also means Put
// performs no allocation, which boxing a bare slice header would (SA6002).
var pathCachePool = sync.Pool{
	New: func() interface{} {
		s := make([]pathResult, 0, 8)
		return &s
	},
}

func getPathCache() *[]pathResult {
	p := pathCachePool.Get().(*[]pathResult)
	*p = (*p)[:0]
	return p
}

func putPathCache(cache *[]pathResult) {
	if cache == nil || cap(*cache) > maxPathCacheCap {
		return
	}
	resetPathCache(cache)
	pathCachePool.Put(cache)
}

// resetPathCache truncates a pooled entry and clears its elements. pathResult
// holds substrings of the caller's path string, so without the clear a
// recycled entry would pin those strings for as long as the entry lives.
// Clearing the whole capacity (clear compiles to a memclr) also covers
// elements past len, in case a caller appended through a local header.
func resetPathCache(cache *[]pathResult) {
	clear((*cache)[:cap(*cache)])
	*cache = (*cache)[:0]
}
