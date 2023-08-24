package zstring

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestRand(T *testing.T) {
	t := zlsgo.NewTest(T)
	t.Log(Rand(4), Rand(10), Rand(4, "a1"))
	t.Log(RandInt(4, 10), RandInt(1, 10), RandInt(1, 2), RandInt(1, 0))
	t.Log(RandUint32Max(10), RandUint32Max(100), RandUint32Max(1000), RandUint32Max(10000))
	t.Log(UUID())
}

func TestUniqueID(T *testing.T) {
	t := zlsgo.NewTest(T)
	t.Log(UniqueID(4), UniqueID(10), UniqueID(0), UniqueID(-6))
}

func TestWeightedRand(T *testing.T) {
	t := zlsgo.NewTest(T)

	c := map[interface{}]uint32{
		"a": 1,
		"b": 6,
		"z": 8,
		"c": 3,
	}

	t.Log(WeightedRand(c))

	w, err := NewWeightedRand(c)
	t.NoError(err)
	t.Log(w.Pick())

	_, err = NewWeightedRand(map[interface{}]uint32{
		"a": ^uint32(0),
		"b": 6,
	})
	t.Log(err)
	t.EqualTrue(err != nil)
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
