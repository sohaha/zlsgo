/*
 * @Author: seekwe
 * @Date:   2019-05-29 15:15:22
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-05-29 15:15:34
 */
package zls

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
		var byteSlice []byte
		byteSlice = make([]byte, 0, BuffSize)
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
