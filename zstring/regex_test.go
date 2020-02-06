package zstring

import (
	"github.com/sohaha/zlsgo"
	"regexp"
	"testing"
)

func TestExtract(T *testing.T) {
	t := zlsgo.NewTest(T)
	res, err := Extract(`abc(\d{2}).*(\w)`, "abc123999ok")
	t.Equal(true, err == nil)
	t.Equal([]string{"12", "k"}, res[1:])
}

func TestRegex(T *testing.T) {
	t := zlsgo.NewTest(T)
	t.Equal(true, IsMatch("是我啊", "这就是我啊!"))
	t.Equal(false, IsMatch("是你呀", "这就是我啊!"))

	phone := "13800138000"
	isPhone := IsMatch(`^1[\d]{10}$`, phone)
	t.Equal(true, isPhone)
	phone = "1380013800x"
	isPhone = IsMatch(`^1[\d]{10}$`, phone)
	t.Equal(false, isPhone)
}

func BenchmarkRegex1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsMatch("是我啊", "这就是我啊!")
	}
}

func BenchmarkRegex2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r, _ := regexp.Compile("是我啊")
		r.Match(String2Bytes("这就是我啊!"))
	}
}

func BenchmarkRegex3(b *testing.B) {
	r, _ := regexp.Compile("是我啊")
	for i := 0; i < b.N; i++ {
		r.Match(String2Bytes("这就是我啊!"))
	}
}
