package zutil

import (
	"bytes"
	"sync"
)

const BuffSize = 10 * 1024

var buffPool sync.Pool

func GetBuff() *bytes.Buffer {
	var buffer *bytes.Buffer
	item := buffPool.Get()
	if item == nil {
		byteSlice := make([]byte, 0, BuffSize)
		buffer = bytes.NewBuffer(byteSlice)
	} else {
		buffer = item.(*bytes.Buffer)
	}
	return buffer
}

func PutBuff(buffer *bytes.Buffer) {
	buffer.Reset()
	buffPool.Put(buffer)
}
