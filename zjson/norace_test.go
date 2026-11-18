//go:build !race

package zjson

// raceDetector reports whether the race detector is active. Allocation counts
// are inflated by race instrumentation, so allocation assertions only run in
// non-race builds.
const raceDetector = false
