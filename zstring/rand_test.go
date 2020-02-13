package zstring

import (
	"github.com/sohaha/zlsgo"
	"testing"
)

func TestRand(T *testing.T) {
	t := zlsgo.NewTest(T)
	t.Log(Rand(4))
	t.Log(Rand(10))
	t.Log(Rand(4, "a1"))
	t.Log(RandInt(4, 10))
}

func BenchmarkRandStr(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Rand(1)
	}
}

func BenchmarkRandStr2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Rand(10)
	}
}

func BenchmarkRandInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandInt(10, 99)
	}
}

func BenchmarkRandInt2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandInt(10000, 99999)
	}
}
