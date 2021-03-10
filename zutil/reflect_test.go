package zutil

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
)

type TestSt struct {
	Name string
	I    int `z:"iii"`
}

func (*TestSt) RunTest(t *testing.T) {
	t.Log("RunTest")
}

func (*TestSt) RunTest2() {}

type TestSt2 struct {
	Name  string
	Test2 bool
}

func (tt *TestSt2) RunTest(t *testing.T) {
	t.Log("RunTest", tt.Name)
}

func TestRunAllMethod(t *testing.T) {
	tt := zlsgo.NewTest(t)
	err := RunAllMethod(&TestSt{}, t)
	t.Log(err)
	tt.Equal(true, err != nil)

	err = RunAllMethod(&TestSt2{Name: "AllMethod"}, t)
	t.Log(err)
	tt.Equal(true, err == nil)

	err = RunAssignMethod(&TestSt2{Name: "AssignMethod"}, func(methodName string) bool {
		t.Log("methodName:", methodName)
		return true
	}, t)
	t.Log(err)
	tt.Equal(true, err == nil)
}

func TestGetAllMethod(t *testing.T) {
	tt := zlsgo.NewTest(t)
	err := GetAllMethod(&TestSt{}, func(numMethod int, m reflect.Method) error {
		t.Log(m.Name)
		if m.Name != "RunTest" && m.Name != "RunTest2" {
			return errors.New("mismatch")
		}
		return nil
	})
	tt.Equal(true, err == nil)

	err = GetAllMethod("test", nil)
	t.Log(err)
	// tt.Equal(true, err != nil)

	err = GetAllMethod(&TestSt{}, nil)
	t.Log(err)
	// tt.Equal(true, err == nil)
}

func TestReflectStructField(t *testing.T) {
	tt := zlsgo.NewTest(t)
	var test = &TestSt{}
	tf := reflect.TypeOf(test)
	// fieldPtr := uintptr(unsafe.Pointer(test))
	err := ReflectStructField(tf, func(numField int, fieldTag string,
		field reflect.StructField) error {
		// fieldPtrOffset := fieldPtr + field.Offset
		switch field.Type.Kind() {
		case reflect.String:
			// noinspection GoVetUnsafePointer
			// *((*string)(unsafe.Pointer(fieldPtrOffset))) = "ok"
		}
		return nil
	})
	tt.EqualNil(err)
	t.Log(test)
}

func TestReflectForNumField(t *testing.T) {
	tt := zlsgo.NewTest(t)
	var test = &struct {
		TestSt
		*TestSt2
		New       bool
		UpdatedAt time.Time
		Updated   uint8
		T2p       *TestSt2
		T2        TestSt2
	}{}
	rv := reflect.ValueOf(test)
	rv = rv.Elem()
	err := ReflectForNumField(rv, func(fieldName, fieldTag string, kind reflect.Kind, field reflect.Value) error {
		t.Log(fieldTag, kind, field.Kind())
		return nil
	})
	tt.EqualNil(err)
}

func TestNonzero(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	v := reflect.ValueOf(0)
	t.EqualTrue(!Nonzero(v))

	v = reflect.ValueOf(10.00)
	t.EqualTrue(Nonzero(v))

	v = reflect.ValueOf(false)
	t.EqualTrue(!Nonzero(v))
}

func TestCanInline(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	v := reflect.TypeOf(0)
	t.EqualTrue(CanInline(v))

	v = reflect.TypeOf(&TestSt{})
	t.EqualTrue(!CanInline(v))

	v = reflect.TypeOf(TestSt{Name: "yes"})
	t.EqualTrue(CanInline(v))

	v = reflect.TypeOf(map[string]interface{}{"d": 10, "a": "zz"})
	t.EqualTrue(!CanInline(v))

	v = reflect.TypeOf([...]int{10, 256})
	t.EqualTrue(CanInline(v))

	v = reflect.TypeOf(func() {})
	t.EqualTrue(!CanInline(v))

}

func TestSetValue(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	t.Log(666)
	vv := &TestSt2{Name: "1"}

	v := reflect.ValueOf(vv)
	err := ReflectForNumField(v.Elem(), func(fieldName, fieldTag string,
		kind reflect.Kind, field reflect.Value) error {
		if fieldName == "Test2" {
			tt.Log(fieldName, true)
			return SetValue(kind, field, true)
		}
		tt.Log(fieldName, "new")
		return SetValue(kind, field, "new")
	})
	t.EqualNil(err)
	t.Equal("new", vv.Name)
	t.Equal(true, vv.Test2)
}
