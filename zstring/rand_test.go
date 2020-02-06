package zstring

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestRand(T *testing.T) {
	t := zlsgo.NewTest(T)
	t.Log(Rand(4))
	t.Log(Rand(10))
	t.Log(Rand(4, "a1"))
	t.Log(RandInt(4, 10))
}

func BenchmarkRand(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Rand(10)
	}
}
