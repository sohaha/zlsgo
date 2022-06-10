package ztype_test

import (
	"reflect"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/ztype"
)

func TestNewStruct(t *testing.T) {
	tt := zlsgo.NewTest(t)

	s := ztype.NewStruct()
	s.AddField("Name", "")
	s.AddField("Age", reflect.TypeOf(1), `json:"age"`)
	data := `{"Name":"test","age":33}`
	v := s.Interface()
	err := zjson.Unmarshal(data, v)
	tt.NoError(err)
	t.Logf("%+v\n", v)

	childS := ztype.NewStruct()
	childS.AddField("Name", "")
	childS.AddField("Child", s, `json:"child"`)
	v2 := childS.Interface()
	data2 := `{"Name":"testChild","child":` + data + `}`
	err = zjson.Unmarshal(data2, v2)
	tt.NoError(err)
	t.Logf("%+v\n", v2)
}

func TestNewSliceStruct(t *testing.T) {
	tt := zlsgo.NewTest(t)

	s := ztype.NewSliceStruct()
	s.AddField("Name", "")
	s.AddField("Age", reflect.TypeOf(1), `json:"age"`)
	data := `[{"Name":"test","age":33},{"Name":"test2","age":100}]`
	v := s.Interface()
	err := zjson.Unmarshal(data, v)
	tt.NoError(err)
	t.Logf("%+v %s\n", v, data)

	childS := ztype.NewSliceStruct()
	childS.AddField("Name", "")
	childS.AddField("Child", s, `json:"child"`)
	v2 := childS.Interface()
	data2 := `[{"Name":"testChild","child":` + data + `}]`
	err = zjson.Unmarshal(data2, v2)
	tt.NoError(err)
	t.Logf("%+v %s\n", v2, data2)
}

func TestNewMapStruct(t *testing.T) {
	tt := zlsgo.NewTest(t)

	s := ztype.NewMapStruct("")
	s.AddField("Name", "")
	s.AddField("Age", reflect.TypeOf(1), `json:"age"`)
	data := `{"test1":{"Name":"11","age":33},"test2":{"Name":"22","age":99}}`
	v := s.Interface()
	err := zjson.Unmarshal(data, v)
	tt.NoError(err)
	t.Logf("%+v %s\n", v, data)

	childS := ztype.NewMapStruct("")
	childS.AddField("Name", "")
	childS.AddField("Child", s, `json:"child"`)
	v2 := childS.Interface()
	data2 := `{"test1":{"Name":"testChild","child":` + data + `}}`
	err = zjson.Unmarshal(data2, v2)
	tt.NoError(err)
	t.Logf("%+v %s\n", v2, data2)
}

type testStruct struct {
	Name string
}

func (t *testStruct) Alias() string {
	return "Alias_" + t.Name
}
