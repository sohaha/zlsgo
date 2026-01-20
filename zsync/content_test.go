package zsync

import (
	"context"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
)

func TestMergeContext(t *testing.T) {
	tt := zlsgo.NewTest(t)

	{
		ctx1 := context.Background()
		ctx2 := context.WithValue(context.Background(), "key", "value2")
		ctx3, cancel3 := context.WithCancel(context.Background())

		ctx := MergeContext(ctx1, ctx2, ctx3)
		tt.Equal(reflect.TypeOf(ctx), reflect.TypeOf(&mergeContext{}))
		tt.Equal("value2", ctx.Value("key"))

		now := time.Now()
		go func() {
			time.Sleep(time.Second / 5)
			cancel3()
		}()

		<-ctx.Done()

		tt.EqualTrue(time.Since(now).Seconds() > 0.2)
	}

	{
		ctx1 := context.Background()
		ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second/5)
		defer cancel2()
		ctx := MergeContext(ctx1, ctx2)
		now := time.Now()
		select {
		case <-ctx.Done():
			tt.EqualTrue(ctx.Err() != nil)
		}
		tt.EqualTrue(time.Since(now).Seconds() > 0.2)
	}

	{
		ctx1 := context.Background()
		ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second/5)
		ctx3, cancel3 := context.WithTimeout(context.Background(), time.Second/10)
		defer cancel2()
		defer cancel3()
		ctx := MergeContext(ctx1, ctx2, ctx3)
		now := time.Now()
		select {
		case <-ctx.Done():
			tt.EqualTrue(ctx.Err() != nil)
		}
		tt.EqualTrue(time.Since(now).Seconds() > 0.1)
	}
}

func TestMergeContextNoGoroutineLeak(t *testing.T) {
	base := runtime.NumGoroutine()
	for i := 0; i < 20; i++ {
		ctx1, cancel1 := context.WithCancel(context.Background())
		ctx2, cancel2 := context.WithCancel(context.Background())
		ctx3, cancel3 := context.WithCancel(context.Background())
		merged := MergeContext(ctx1, ctx2, ctx3)
		cancel2()
		<-merged.Done()
		cancel1()
		cancel3()
	}
	time.Sleep(200 * time.Millisecond)
	runtime.GC()
	time.Sleep(200 * time.Millisecond)
	after := runtime.NumGoroutine()
	if after > base+10 {
		t.Fatalf("goroutine leak: base=%d after=%d", base, after)
	}
}

func TestMergeContextNoLeakWithoutCancelable(t *testing.T) {
	base := runtime.NumGoroutine()
	for i := 0; i < 50; i++ {
		_ = MergeContext(context.Background(), context.TODO())
	}
	time.Sleep(200 * time.Millisecond)
	runtime.GC()
	time.Sleep(200 * time.Millisecond)
	after := runtime.NumGoroutine()
	if after > base+10 {
		t.Fatalf("goroutine leak: base=%d after=%d", base, after)
	}
}
