package zutil

import (
	"errors"
	"reflect"
	"testing"

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

type TestSt2 struct{}

func (*TestSt2) RunTest(t *testing.T) {
	t.Log("RunTest")
}

func TestRunAllMethod(t *testing.T) {
	tt := zlsgo.NewTest(t)
	err := RunAllMethod(&TestSt{}, t)
	t.Log(err)
	tt.Equal(true, err != nil)

	err = RunAllMethod(&TestSt2{}, t)
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
