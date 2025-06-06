package ztime

import (
	"context"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
)

func TestSleep(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Run("base", func(tt *zlsgo.TestUtil) {
		tt.Parallel()
		now := time.Now()
		Sleep(context.Background(), time.Second/5)
		tt.EqualTrue(time.Since(now) >= time.Second/5)
	})

	tt.Run("cancel", func(tt *zlsgo.TestUtil) {
		tt.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second/5)
		defer cancel()
		now := time.Now()
		Sleep(ctx, time.Second)
		tt.EqualTrue(time.Since(now) >= time.Second/5)
	})
}
