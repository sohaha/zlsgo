package zerror_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zerror"
)

func TestTryCatch(t *testing.T) {
	tt := zlsgo.NewTest(t)
	err := zerror.TryCatch(func() error {
		zerror.Panic(zerror.New(500, "测试"))
		return nil
	})
	tt.EqualTrue(err != nil)
	tt.Equal("测试", err.Error())
	code, _ := zerror.UnwrapCode(err)
	tt.Equal(zerror.ErrCode(500), code)
}

func BenchmarkTryCatch_normal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = zerror.TryCatch(func() error {
			e := strconv.Itoa(i)
			_ = e
			return nil
		})
	}
}

func BenchmarkTryCatch_panic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = zerror.TryCatch(func() error {
			e := strconv.Itoa(i)
			panic(e)
			return nil
		})
	}
}

func BenchmarkTryCatch_error(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = zerror.TryCatch(func() error {
			e := strconv.Itoa(i)
			return errors.New(e)
		})
	}
}
