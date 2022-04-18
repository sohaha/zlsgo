package zvalid

import (
	"strings"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
)

func TestValidRule(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	v := Text("a1Cb.1").Required().HasLetter().HasLower().HasUpper().HasNumber().HasSymbol().HasString("b").HasPrefix("a").HasSuffix("1").Password().StrongPassword()
	err := v.Error()
	t.Log(err)
	t.Equal(true, v.Ok())
}

func TestValidNew(tt *testing.T) {
	var err error
	var str string
	t := zlsgo.NewTest(tt)

	valid := New()
	validObj := valid
	err = valid.Error()
	t.Equal(ErrNoValidationValueSet, err)
	tt.Log(str, err)

	validObj.SetAlias("测试 1")
	verifi := validObj.Verifi("test1", "测试1")
	v := verifi.valid()

	t.Equal(nil, v.err)
	tt.Log(v.value, v.err)

	test2 := validObj.Verifi("", "测试2").Required("Test2")
	tt.Log("test2 queue", test2.queue.Len())
	v = test2.valid()
	t.Equal(true, v.Error() != nil)
	tt.Log(v.Value(), err)

	v = valid.valid()
	t.Equal(ErrNoValidationValueSet, v.Error())
	tt.Log(v.value, v.err, v.setRawValue)

	test3 := validObj.IsNumber().Verifi("test3", "测试3")
	v = test3.valid()
	tt.Log("test3 queue", test3.queue.Len())
	t.Equal(true, v.Error() != nil)
	t.Equal(v.value, test3.Value())
	t.Equal(v.Error(), test3.Error())
	tt.Log(v.value, v.err)
	tt.Log(test3.Value(), test3.Error())

	test4 := validObj.Verifi("", "测试4").Customize(func(rawValue string, err error) (newValue string, newErr error) {
		newValue = "test4"
		tt.Log("重置值")
		return
	})

	str, err = test4.String()
	tt.Log("test4 queue", test4.queue.Len())
	t.Equal(nil, err)
	tt.Log(str, err)

}

func TestInt(t *testing.T) {
	tt := zlsgo.NewTest(t)
	i, err := Int(64).MaxInt(60).Int()
	tt.Equal(true, err != nil)
	t.Log(err)
	t.Log(i)

	i, err = Int(6).MaxInt(60).Int()
	tt.EqualNil(err)
	t.Log(i)
}

func TestFloat64(t *testing.T) {
	tt := zlsgo.NewTest(t)
	i, err := Int(6).MaxInt(60).Float64()
	tt.EqualNil(err)
	t.Log(i)
}

func TestBool(t *testing.T) {
	tt := zlsgo.NewTest(t)
	b, err := Text("true").Bool()
	tt.EqualNil(err)
	tt.Equal(true, b)
	b, err = Text("0").Bool()
	tt.EqualNil(err)
	tt.Equal(false, b)
}

func TestSplit(t *testing.T) {
	tt := zlsgo.NewTest(t)
	b, err := Text("1,2,3,4,5").Split(",")
	tt.EqualNil(err)
	t.Log(b)
	tt.Equal(5, len(b))
}

func TestPassword(t *testing.T) {
	var err error
	tt := zlsgo.NewTest(t)

	err = New().Verifi("123aA.").Password().Error()
	tt.EqualNil(err)
	err = New().Verifi("a", "pass2").Password().Error()
	tt.Equal(true, err != nil)
	tt.Log(err)
	err = New().Verifi("").Password().Error()
	tt.Equal(true, err != nil)
	tt.Log(err)

	err = New().Verifi("123aA.").StrongPassword().Error()
	tt.EqualNil(err)
	err = New().Verifi("123aA").StrongPassword().Error()
	tt.Equal(true, err != nil)
	err = New().Verifi("").StrongPassword().Error()
	tt.Equal(true, err != nil)

	now := time.Now()
	str, err := New().Verifi(`123`).EncryptPassword().String()
	tt.EqualNil(err)
	t.Log(str)
	t.Log("time", time.Since(now).Seconds())

	now = time.Now()
	str, err = New().Verifi(`123`).CheckPassword(str).String()
	tt.EqualNil(err)
	t.Log(str)
	t.Log("time", time.Since(now).Seconds())

	now = time.Now()
	str, err = New().Verifi(`1231`).CheckPassword(str).String()
	t.Log(str, err)
	t.Log("time", time.Since(now).Seconds())

	now = time.Now()
	str, err = New().Verifi(`123`).EncryptPassword(14).String()
	tt.EqualNil(err)
	t.Log(str)
	t.Log("time", time.Since(now).Seconds())

	now = time.Now()
	str, err = New().Verifi(`123`).EncryptPassword(4).String()
	tt.EqualNil(err)
	t.Log(str)
	t.Log("time", time.Since(now).Seconds())
}

func BenchmarkStr(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := "test"
		if !strings.HasPrefix(s, "t") {
			b.Fatal()
		}
	}
}

func BenchmarkStr2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := "test"
		if ok := Text(s).HasPrefix("t").Ok(); !ok {
			b.Fatal()
		}
	}
}
