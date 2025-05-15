//go:build go1.10
// +build go1.10

package zstring

import (
	"strings"
)

// Buffer creates a new strings.Builder with optional initial capacity.
// This implementation uses the more efficient strings.Builder available in Go 1.10+.
func Buffer(size ...int) *strings.Builder {
	var b strings.Builder
	if len(size) > 0 {
		b.Grow(size[0])
	}
	return &b
}
