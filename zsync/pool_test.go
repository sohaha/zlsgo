//go:build go1.18
// +build go1.18

package zsync

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestNewPool(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Run("base", func(tt *zlsgo.TestUtil) {
		pool := NewPool(func() int {
			return 1
		})
		tt.Equal(1, pool.Get())
	})
}
