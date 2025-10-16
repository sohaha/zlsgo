//go:build go1.18
// +build go1.18

package zsync

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestNewValue(t *testing.T) {
	tt := zlsgo.NewTest(t)

	arr := []string{"1", "bool", "test", "???"}

	tt.Run("base", func(tt *zlsgo.TestUtil) {
		var wg WaitGroup
		var v AtomicValue[string]
		for i := range arr {
			b := arr[i]
			wg.GoTry(func() {
				t.Log("-", v.Load())
				v.Store(b)
				t.Log("=", v.Load())
			})
		}

		tt.NoError(wg.Wait())
	})

	tt.Run("new", func(tt *zlsgo.TestUtil) {
		var wg WaitGroup
		v := NewValue("xxx")
		for i := range arr {
			b := arr[i]
			wg.GoTry(func() {
				t.Log("-", v.Load())
				v.Store(b)
				t.Log("=", v.Load())
			})
		}

		tt.NoError(wg.Wait())
	})

	tt.Run("more", func(tt *zlsgo.TestUtil) {
		var v AtomicValue[string]

		tt.Equal("", v.Load(), true)

		v.Store("yyy")
		tt.Equal("yyy", v.Load(), true)

		tt.Equal(false, v.CAS("x1", "x2"))
		tt.Equal(true, v.CAS("yyy", "x2"))
		tt.Equal("x2", v.Load(), true)

		tt.Equal("x2", v.Swap("zzz"), true)
		tt.Equal("zzz", v.Load(), true)
	})

	tt.Run("empty", func(tt *zlsgo.TestUtil) {
		var v AtomicValue[string]
		old := v.Swap("xxx")
		tt.Equal("", old)
		tt.Equal("xxx", v.Load(), true)
	})
}

func TestAtomicValuePointer(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var v AtomicValue[*int]
	var zero *int
	p1 := new(int)
	p2 := new(int)

	tt.Equal((*int)(nil), v.Load())

	ok := v.CAS(zero, p1)
	tt.Equal(false, ok)

	v.Store(p1)
	tt.Equal(p1, v.Load())

	ok = v.CAS(p1, p2)
	tt.Equal(true, ok)
	tt.Equal(p2, v.Load())

	ok = v.CAS(p1, p2)
	tt.Equal(false, ok)
}

func TestAtomicValueSlice(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var v AtomicValue[[]int]
	s1 := []int{1, 2, 3}
	s2 := []int{4, 5}

	tt.Equal(([]int)(nil), v.Load())

	v.Store(s1)
	tt.Equal(s1, v.Load())

	old := v.Swap(s2)
	tt.Equal(s1, old)
	tt.Equal(s2, v.Load())

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on CAS with slice type")
		}
	}()
	_ = v.CAS(s2, s1)
}

func TestAtomicValueMap(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var v AtomicValue[map[string]int]
	m1 := map[string]int{"a": 1}
	m2 := map[string]int{"b": 2}

	tt.Equal((map[string]int)(nil), v.Load())

	v.Store(m1)
	tt.Equal(m1, v.Load())

	old := v.Swap(m2)
	tt.Equal(m1, old)
	tt.Equal(m2, v.Load())

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on CAS with map type")
		}
	}()
	_ = v.CAS(m2, m1)
}

func TestAtomicValueChan(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var v AtomicValue[chan int]
	c1 := make(chan int)
	c2 := make(chan int)
	c3 := make(chan int)

	v.Store(c1)
	tt.Equal(c1, v.Load())

	ok := v.CAS(c1, c2)
	tt.Equal(true, ok)
	tt.Equal(c2, v.Load())

	ok = v.CAS(c1, c3)
	tt.Equal(false, ok)
}

func TestAtomicValueFunc(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var v AtomicValue[func() int]
	f1 := func() int { return 1 }
	f2 := func() int { return 2 }

	v.Store(f1)
	tt.Equal(1, v.Load()())

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on CAS with func type")
		}
	}()
	_ = v.CAS(f1, f2)
}

func TestAtomicValueAny(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var v AtomicValue[any]

	v.Store(1)
	tt.Equal(1, v.Load())

	ok := v.CAS(2, 3)
	tt.Equal(false, ok)
	tt.Equal(1, v.Load())

	v.Store("x")
	ok = v.CAS("x", "y")
	tt.Equal(true, ok)
	tt.Equal("y", v.Load())

	v.Store([]int{1})
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on CAS with any holding slice")
		}
	}()
	_ = v.CAS([]int{1}, []int{2})
}

func TestAtomicValueEmptySwapReturnsZero(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var v AtomicValue[int]
	old := v.Swap(10)
	tt.Equal(0, old)
	tt.Equal(10, v.Load())

	var vp AtomicValue[*int]
	p := new(int)
	oldPtr := vp.Swap(p)
	tt.Equal((*int)(nil), oldPtr)
	tt.Equal(p, vp.Load())
}
