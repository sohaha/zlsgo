package zsync

import (
	"context"
	"reflect"
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
