package zsync

import (
	"sync"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zutil"
)

func TestWaitGroup(t *testing.T) {
	t.Run("base", func(t *testing.T) {
		tt := zlsgo.NewTest(t)
		t.Parallel()
		count := zutil.NewInt64(0)
		var wg WaitGroup
		for i := 0; i < 100; i++ {
			wg.Go(func() {
				count.Add(1)
			})
		}
		err := wg.Wait()
		tt.NoError(err)
		tt.Equal(int64(100), count.Load())
	})

	t.Run("err", func(t *testing.T) {
		tt := zlsgo.NewTest(t)
		t.Parallel()
		count := zutil.NewInt64(0)
		var wg WaitGroup
		for i := 0; i < 100; i++ {
			var ii = i
			wg.GoTry(func() {
				count.Add(1)
				if ii > 0 && ii%5 == 0 {
					panic("manual panic")
				}
			})
		}
		err := wg.Wait()
		tt.EqualTrue(err != nil)
		t.Logf("%+v", err)
		zlog.Error(err)
		tt.Equal(int64(100), count.Load())
	})
}

func BenchmarkWaitGroup_Go(b *testing.B) {
	var wg sync.WaitGroup
	var count int64
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			count++
			wg.Done()
		}()
	}
	wg.Wait()
}
