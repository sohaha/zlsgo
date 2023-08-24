package zreflect_test

import (
	"reflect"
	"testing"

	"github.com/sohaha/zlsgo/zreflect"
)

func BenchmarkZReflect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v := zreflect.NewType(zreflect.Demo)
		_ = v.NumMethod()
	}
}

func BenchmarkZReflectRaw(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v := zreflect.NewType(zreflect.Demo)
		_ = v.Native().NumMethod()
	}
}

func BenchmarkGReflect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v := reflect.TypeOf(zreflect.Demo)
		_ = v.NumMethod()
	}
}
