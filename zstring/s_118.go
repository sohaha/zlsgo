//go:build !go1.10
// +build !go1.10

package zstring

import "bytes"

// Buffer Buffer
func Buffer(size ...int) *bytes.Buffer {
	b := bytes.NewBufferString("")
	return &b
}
