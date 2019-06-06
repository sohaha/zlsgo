//+build !go1.10

package zstring

import "bytes"

// Buffer Buffer
func Buffer() bytes.Buffer {
	return bytes.NewBufferString("")
}
