package zvalid

import (
	"github.com/sohaha/zlsgo"
	"testing"
)

func TestValidNew(tt *testing.T) {
	var err error
	var str string
	t := zlsgo.NewTest(tt)

	valid := New()
	validObj := clone(valid)
	err = valid.Error()
	t.Equal(ErrNoValidationValueSet, err)
	tt.Log(str, err)

	str, err = validObj.Verifi("test1", "测试1").Result()
	t.Equal(nil, err)
	tt.Log(str, err)

	str, err = validObj.Required().Verifi("", "测试2").Result()
	t.Equal(true, err != nil)
	tt.Log(str, err)

	str, err = valid.Result()
	t.Equal(ErrNoValidationValueSet, err)
	tt.Log(str, err)

	test3 := validObj.IsNumber().Verifi("test3", "测试3")
	str, err = test3.Result()
	t.Equal(true, err != nil)
	t.Equal(str, test3.Value())
	t.Equal(err, test3.Error())
	tt.Log(str, err)
	tt.Log(test3.Value(), test3.Error())

	str, err = validObj.Verifi("", "测试4").Customize(func(rawValue string, err error) (newValue string, newErr error) {
		newValue = "test4"
		return
	}).Result()
	t.Equal(nil, err)
	tt.Log(str, err)
}

func BenchmarkNew0(b *testing.B) {
	n := Text("")
	for i := 0; i < b.N; i++ {
		c := clone(n)
		_ = c
	}
}

func BenchmarkNew0_1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		n := Text("")
		c := clone(n)
		_ = c
	}
}

func BenchmarkNew1(b *testing.B) {
	n := New()
	for i := 0; i < b.N; i++ {
		c := clone(n)
		_ = c
	}
}

func BenchmarkNew1_1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		n := New()
		c := clone(n)
		_ = c
	}
}

func BenchmarkNew2(b *testing.B) {
	n := New()
	for i := 0; i < b.N; i++ {
		c := n
		_ = c
	}
}

func BenchmarkNew2_2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		n := New()
		c := n
		_ = c
	}
}
