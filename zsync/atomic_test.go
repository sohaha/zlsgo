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
