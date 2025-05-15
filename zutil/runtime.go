package zutil

import (
	"bytes"
	"runtime"
	"strconv"
)

// GetGid returns the ID of the current goroutine.
// This is useful for debugging and logging purposes to track which goroutine
// is executing a particular piece of code.
func GetGid() uint64 {
	b := make([]byte, 64)
	runtime.Stack(b, false)
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}
