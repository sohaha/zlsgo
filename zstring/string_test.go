package zstring

import (
	"testing"

	zls "github.com/sohaha/zlsgo"
)

func TestBuffer(T *testing.T) {
	t := zls.NewTest(T)
	l := "0"
	l += "1"
	b := Buffer()
	b.WriteString("0")
	b.WriteString("1")
	t.Equal(l, b.String())
}

func TestRand(T *testing.T) {
	t := zls.NewTest(T)
	t.Log(Rand(4))
	t.Log(Rand(100))
}

func TestLen(T *testing.T) {
	t := zls.NewTest(T)
	s := "我是中国人"
	t.Equal(5, Len(s))
	t.Log(Len(s), len(s))
}

func TestSubstr(T *testing.T) {
	t := zls.NewTest(T)
	s := "0123"
	t.Equal(Substr(s, 1), "123")
	t.Equal(Substr(s, 2, 1), "2")
}

func TestPad(T *testing.T) {
	t := zls.NewTest(T)
	l := "我的这里一共8字"
	t.Equal(8, Len(l))

	s := "我的长度是二十,不够右边补零"
	t.Equal("我的长度是二十,不够右边补零000000", Pad(s, 20, "0", PadRight))

	s2 := "我的长度是二十,不够左边补零"
	t.Equal("000000我的长度是二十,不够左边补零", Pad(s2, 20, "0", PadLeft))

	s3 := "我的长度很长不需要填充"
	t.Equal("我的长度很长不需要填充", Pad(s3, 5, "我的长度很长不需要填充", PadRight))

	t.Equal("长度", Substr(s3, 2, 2))

	s4 := "我的长度是二十,不够两边补零"
	t.Equal("000我的长度是二十,不够两边补零000", Pad(s4, 20, "0", PadSides))
}

func TestFirst(T *testing.T) {
	t := zls.NewTest(T)
	str := "my Name"
	str = Ucfirst(str)
	t.Equal("My Name", str)
	str = Lcfirst(str)
	t.Equal("my Name", str)
}

func BenchmarkStr(b *testing.B) {
	s := ""
	for i := 0; i < b.N; i++ {
		s += "1"
	}
}

func BenchmarkStrBuffer(b *testing.B) {
	s := Buffer()
	for i := 0; i < b.N; i++ {
		s.WriteString("1")
	}
	_ = s.String()
}

func TestTo(T *testing.T) {
	t := zls.NewTest(T)
	s := "我是中国人"
	b := String2Bytes(&s)
	s2 := Bytes2String(b)
	t.Equal(s, s2)
}

func BenchmarkTo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := "我是中国人"
		b := String2Bytes(&s)
		_ = Bytes2String(b)
	}
}

func BenchmarkTo2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := "我是中国人"
		b := []byte(s)
		_ = string(b)
	}
}
