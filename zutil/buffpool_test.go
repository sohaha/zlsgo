package zutil

import (
	"testing"
)

func TestBuff(T *testing.T) {
	buffer := GetBuff()
	PutBuff(buffer)
	buffer = GetBuff()
	buffer.Reset()
}
