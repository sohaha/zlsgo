package zsync

import (
	"runtime"
	_ "unsafe"
)

const (
	cacheLineSize = 64
)

// nextPowOf2 returns the next power of 2 greater than or equal to v.
// This is used internally for sizing data structures that perform better
// with power-of-2 sizes, such as hash tables and lock arrays.
func nextPowOf2(v uint32) uint32 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return v
}

// parallelism returns the effective parallelism level for the current process.
// It returns the minimum of GOMAXPROCS and the number of CPU cores,
// which provides a reasonable estimate of the available concurrent execution capacity.
func parallelism() uint32 {
	maxProcs := uint32(runtime.GOMAXPROCS(0))
	numCores := uint32(runtime.NumCPU())
	if maxProcs < numCores {
		return maxProcs
	}
	return numCores
}
