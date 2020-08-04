package zstring

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestRand(T *testing.T) {
	t := zlsgo.NewTest(T)
	t.Log(Rand(4), Rand(10), Rand(4, "a1"))
	t.Log(RandInt(4, 10), RandInt(1, 10), RandInt(1, 2), RandInt(1, 0))
	t.Log(UUID())
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
