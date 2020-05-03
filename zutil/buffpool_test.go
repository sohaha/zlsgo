package zutil

import (
	"testing"

	"github.com/sohaha/zlsgo/zstring"
)

func TestBuff(T *testing.T) {
	buffer := GetBuff()
	PutBuff(buffer)
	buffer = GetBuff()
	buffer.Reset()
}

func BenchmarkPoolBytes1(b *testing.B) {
	v := []byte("ok")
	for i := 0; i < b.N; i++ {
		var content = GetBuff()
		content.Write(v)
		str := content.Bytes()
		PutBuff(content)
		if string(v) != string(str) {
			b.Fail()
		}
	}
}

func BenchmarkPoolBytes2(b *testing.B) {
	v := []byte("ok")
	for i := 0; i < b.N; i++ {
		var str []byte
		str = []byte("ok")
		if string(v) != string(str) {
			b.Fail()
		}
	}
}

func BenchmarkPoolBytes3(b *testing.B) {
	v := []byte("ok")
	for i := 0; i < b.N; i++ {
		str := zstring.Buffer()
		str.Write(v)
		if string(v) != str.String() {
			b.Fail()
		}
	}
}
