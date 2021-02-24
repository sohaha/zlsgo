package zutil

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zstring"
)

func TestBuff(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	buffer := GetBuff()
	buffer.WriteString("1")
	buffer.WriteString("2")
	t.EqualExit("12", buffer.String())
	PutBuff(buffer)
	t.EqualExit("", buffer.String())
	buffer2 := GetBuff()
	t.EqualExit("", buffer2.String())
	buffer2.Reset()

	buffer3 := GetBuff(0)
	tt.Log(buffer3.Len(), buffer3.Cap())
	PutBuff(buffer3)

	buffer4 := GetBuff(104857609)
	tt.Log(buffer4.Len(), buffer4.Cap())
	PutBuff(buffer4)

	buffer5 := GetBuff(16)
	tt.Log(buffer5.Len(), buffer5.Cap())
	buffer5.WriteString(strings.Repeat("0", 104857609))
	tt.Log(buffer5.Len(), buffer5.Cap())
	buffer5.Reset()
	tt.Log(buffer5.Len(), buffer5.Cap())
}

func BenchmarkPoolBytesPoolMinSize(b *testing.B) {
	v := []byte("ok")
	for i := 0; i < b.N; i++ {
		var content = GetBuff(16)
		content.Write(v)
		str := content.Bytes()
		PutBuff(content)
		if string(v) != string(str) {
			b.Fail()
		}
	}
}

func BenchmarkPoolBytesPoolMaxSize(b *testing.B) {
	v := []byte("ok")
	for i := 0; i < b.N; i++ {
		var content = GetBuff(16)
		content.Write(v)
		str := content.Bytes()
		PutBuff(content)
		if string(v) != string(str) {
			b.Fail()
		}
	}
}

func BenchmarkPoolBytesPool(b *testing.B) {
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
		var str = &bytes.Buffer{}
		str.Write(v)
		if string(v) != str.String() {
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

var sizeString = strings.Repeat("0", 1024*1024)

func BenchmarkPoolBytesPoolMinSize_max(b *testing.B) {
	v := []byte(sizeString)
	s := 16
	for i := 0; i < b.N; i++ {
		var content = GetBuff(s)
		content.Write(v)
		str := content.Bytes()
		PutBuff(content)
		if string(v) != string(str) {
			b.Fail()
		}
	}
}

func BenchmarkPoolBytesPoolMaxSize_max(b *testing.B) {
	v := []byte(sizeString)
	s := 1024 * 1024
	for i := 0; i < b.N; i++ {
		var content = GetBuff(s)
		content.Write(v)
		str := content.Bytes()
		PutBuff(content)
		if string(v) != string(str) {
			b.Fail()
		}
	}
}

func BenchmarkPoolBytesPool_max(b *testing.B) {
	v := []byte(sizeString)
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

func BenchmarkPoolBytes2_max(b *testing.B) {
	v := []byte(sizeString)
	for i := 0; i < b.N; i++ {
		var str = &bytes.Buffer{}
		str.Write(v)
		if string(v) != str.String() {
			b.Fail()
		}
	}
}

func BenchmarkPoolBytes3_max(b *testing.B) {
	v := []byte(sizeString)
	for i := 0; i < b.N; i++ {
		str := zstring.Buffer()
		str.Write(v)
		if string(v) != str.String() {
			b.Fail()
		}
	}
}
