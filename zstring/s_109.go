//go:build !go1.10
// +build !go1.10

package zstring

import (
	"bytes"
)

// Buffer creates a new empty bytes.Buffer.
// This implementation is for Go versions prior to 1.10.
func Buffer(size ...int) *bytes.Buffer {
	b := bytes.NewBufferString("")
	return &b
}
