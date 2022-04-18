package zreflect_test

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zreflect"
)

func TestMapToStruct(t *testing.T) {
	var demo DemoSt
	tt := zlsgo.NewTest(t)
	err := zreflect.MapToStruct(data, &demo)
	t.Log(demo, err)
	tt.EqualNil(err)
	tt.Equal(data["username"], demo.Name)
	tt.Equal(data["Hobby"], demo.Hobby)

	var demo2 struct {
		Name string `json:"username"`
		Age  int
	}
	err = zreflect.MapToStruct(data, &demo2)
	t.Log(demo2, err)
}

func BenchmarkMapToStruct(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var demo DemoSt
		_ = zreflect.MapToStruct(data, &demo)
	}
}
