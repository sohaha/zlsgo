package zreflect

import (
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
)

var Demo = DemoSt{Name: "test reflect", Remark: "yes", Date2: time.Now(), pri: "is private"}

func TestGetUnexportedField(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Run("pub", func(tt *zlsgo.TestUtil) {
		value, err := GetUnexportedField(ValueOf(Demo), "Name")
		tt.Log(value, err)
		tt.NoError(err)
		tt.EqualExit("test reflect", value.(string))

		value, err = GetUnexportedField(ValueOf(&Demo), "Name")
		tt.NoError(err)
		tt.Log(value, err)
		tt.EqualExit("test reflect", value.(string))
	})

	tt.Run("pri", func(tt *zlsgo.TestUtil) {
		value, err := GetUnexportedField(ValueOf(Demo), "pri")
		tt.Log(value, err)
		tt.EqualTrue(err != nil)
		tt.EqualExit(nil, value)

		value, err = GetUnexportedField(ValueOf(&Demo), "pri")
		tt.Log(value, err)
		tt.NoError(err)
		tt.EqualExit("is private", value.(string))
	})

	tt.Run("not exists", func(tt *zlsgo.TestUtil) {
		value, err := GetUnexportedField(ValueOf(Demo), "pri_not_exists")
		tt.Log(value, err)
		tt.EqualTrue(err != nil)
		tt.EqualExit(nil, value)

		value, err = GetUnexportedField(ValueOf(&Demo), "pri_not_exists")
		tt.Log(value, err)
		tt.EqualTrue(err != nil)
		tt.EqualExit(nil, value)
	})
}

func TestSetUnexportedField(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Run("pub", func(tt *zlsgo.TestUtil) {
		v := Demo
		err := SetUnexportedField(ValueOf(&v), "Name", "new name")
		tt.NoError(err)
		tt.EqualExit("new name", v.Name)

		v = Demo
		err = SetUnexportedField(ValueOf(v), "Name", "new name")
		tt.Log(err)
		tt.EqualTrue(err != nil)
		tt.EqualExit("test reflect", v.Name)

		v = Demo
		err = SetUnexportedField(ValueOf(&v), "Name", 1)
		tt.Log(err)
		tt.EqualTrue(err != nil)

		v = Demo
		err = SetUnexportedField(ValueOf(&v), "Any", 1)
		tt.Log(err)
		tt.NoError(err)
		tt.EqualExit(1, v.Any)
	})

	tt.Run("pri", func(tt *zlsgo.TestUtil) {
		v := Demo
		err := SetUnexportedField(ValueOf(&v), "pri", "new pri")
		tt.NoError(err)
		tt.EqualExit("new pri", v.pri)

		v = Demo
		err = SetUnexportedField(ValueOf(v), "pri", "new name")
		tt.Log(err)
		tt.EqualTrue(err != nil)
		tt.EqualExit("is private", v.pri)

		v = Demo
		err = SetUnexportedField(ValueOf(&v), "pri", 1)
		tt.Log(err)
		tt.EqualTrue(err != nil)

		v = Demo
		err = SetUnexportedField(ValueOf(&v), "any", 1)
		tt.Log(err)
		tt.NoError(err)
		tt.EqualExit(1, v.any)
	})

	tt.Run("not exists", func(tt *zlsgo.TestUtil) {
		v := Demo
		err := SetUnexportedField(ValueOf(&v), "pri_not_exists", "new pri")
		tt.Log(err)
		tt.EqualTrue(err != nil)

		v = Demo
		err = SetUnexportedField(ValueOf(v), "pri", "new name")
		tt.Log(err)
		tt.EqualTrue(err != nil)
	})
}
