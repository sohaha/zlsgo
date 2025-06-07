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

	v, err := WeightedRand(map[interface{}]uint32{"1": 0})
	t.Log(v, err)
	t.EqualTrue(err == nil)
}

func TestNewNanoID(T *testing.T) {
	t := zlsgo.NewTest(T)
	t.Log(NewNanoID(10))
	t.Log(NewNanoID(10))
	t.Log(NewNanoID(10, "1234"))
	t.Log(NewNanoID(10, "1234"))
}

func BenchmarkNanoID(b *testing.B) {
	b.Run("Nano", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			NewNanoID(21)
		}
	})

	b.Run("Rand", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Rand(21)
		}
	})

	b.Run("NanoID-10", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			NewNanoID(10)
		}
	})

	b.Run("NanoID-21", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			NewNanoID(21)
		}
	})

	b.Run("NanoID-50", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			NewNanoID(50)
		}
	})

	b.Run("NanoID-ASCII", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			NewNanoID(21, "0123456789abcdefghijklmnopqrstuvwxyz")
		}
	})

	b.Run("NanoID-NonASCII", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			NewNanoID(21, "0123456789абвгдеёжзийклмнопрстуфхцчшщъыьэюя")
		}
	})
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

func TestUUID(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Run("base", func(tt *zlsgo.TestUtil) {
		for i := 0; i < 10; i++ {
			tt.Log(UUID())
		}
	})

	tt.Run("unique", func(tt *zlsgo.TestUtil) {
		const count = 1_000_000
		uuids := make(map[string]bool, count)

		for i := 0; i < count; i++ {
			u := UUID()
			if uuids[u] {
				tt.Fatal("Duplicate UUID:", u)
			}
			uuids[u] = true
		}

		tt.Equal(count, len(uuids))
	})
}
