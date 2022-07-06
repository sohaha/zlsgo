package zreflect_test

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zreflect"
)

func TestStruct(t *testing.T) {
	var demo DemoSt
	tt := zlsgo.NewTest(t)
	err := zreflect.Map2Struct(data, &demo)
	t.Log(demo, err)
	tt.EqualNil(err)
	tt.Equal(data["username"], demo.Name)
	tt.Equal(data["Hobby"], demo.Hobby)

	var demo2 struct {
		Name string `json:"username"`
		Age  int
	}
	err = zreflect.Map2Struct(data, &demo2)
	t.Log(demo2, err)

	m, err := zreflect.Struct2Map(demo)
	tt.NoError(err)
	t.Logf("%+v\n", m)
}

func BenchmarkMap2Struct(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var demo DemoSt
		_ = zreflect.Map2Struct(data, &demo)
	}
}
