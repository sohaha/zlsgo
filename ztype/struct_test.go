package ztype_test

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/ztime"
	"github.com/sohaha/zlsgo/ztype"
)

type demo struct {
	Name   string
	Remark string `json:"msg"`
	Age    int    `json:"a"`
}

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

	d := demo{Name: "test", Age: 33, Remark: "msg"}
	err = zjson.Unmarshal(data, &d)
	tt.NoError(err)
	t.Logf("%+v\n", d)

	d2 := ztype.NewStruct()
	d2.Merge(d)
	d2.RemoveField("Remark")
	d2.Interface()

	t.Logf("%+v\n", d2.String())
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

type Rows1 struct {
	DataDate    ztime.LocalTime `json:"data_date"`
	CreatedDate time.Time       `json:"created_at"`
	Name        string          `json:"name"`
	DeletedAt   int
	Age         int8 `json:"age"`
}

type Rows2 struct {
	Name string `json:"name2"`
	Key  string `json:"key"`
	Rows1
	DeletedAt int `json:"deleted_at"`
}

func TestStructEdit(t *testing.T) {
	tt := zlsgo.NewTest(t)

	now := time.Now()
	r1 := Rows1{Name: "R1", DataDate: ztime.LocalTime{Time: now}, CreatedDate: now, Age: 18, DeletedAt: int(time.Now().Unix())}
	j1, _ := json.Marshal(r1)
	tt.Log(string(j1))

	r2 := Rows2{Rows1: r1, Key: "key"}
	j2, _ := json.Marshal(r2)
	tt.Log(string(j2))

	b, _ := ztype.NewStructFromValue(Rows1{})
	b.AddField("DeletedAt", 10, `json:"deleted_at"`)
	b.AddField("Key", "", `json:"key"`)
	b.AddField("Name", "", `json:"name2"`)
	r3 := b.Interface()
	ztype.To(r1, &r3, func(c *ztype.Conver) {
		c.ConvHook = func(name string, i reflect.Value, o reflect.Type) (reflect.Value, bool) {
			if name == "DataDate" {
				tt.Log(name, i, o)
				return i, false
			}
			return i, true
		}
		c.IgnoreTagName = true
	})
	j3, _ := json.Marshal(r3)
	tt.Log(string(j3))
}

func TestStructBuilderHelpers(t *testing.T) {
	b := ztype.NewStruct()
	b.AddField("A", 1, `json:"a"`).AddField("B", "x")
	if !b.HasField("A") || b.GetField("A") == nil {
		t.Fatal("expected field A to exist")
	}
	names := b.FieldNames()
	if len(names) != 2 {
		t.Fatalf("unexpected field names: %#v", names)
	}

	f := b.GetField("A")
	f.SetType(reflect.TypeOf(""))
	f.SetTag(`json:"aa"`)
	_ = b.Interface()

	s := ztype.NewSliceStruct()
	beforeKind := s.Type().Kind()
	other := ztype.NewStruct().AddField("C", 1)
	s.Copy(other)
	if s.Type().Kind() != beforeKind || !s.HasField("C") {
		t.Fatal("Copy did not preserve type or fields")
	}
}

func TestGetFieldNotExists(t *testing.T) {
	b := ztype.NewStruct()
	b.AddField("A", 1)
	if b.GetField("NOPE") != nil {
		t.Fatalf("expected nil for non-existing field")
	}
}

func TestNewMapStructEdgeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

	builder := ztype.NewMapStruct("")
	builder.AddField("Name", "")
	builder.AddField("Age", 0)
	tt.NotNil(builder.Interface())

	builder2 := ztype.NewMapStruct(reflect.TypeOf(""))
	builder2.AddField("Field1", "value")
	tt.NotNil(builder2.Interface())

	sliceBuilder := ztype.NewSliceStruct()
	sliceBuilder.AddField("ID", 0)
	sliceBuilder.AddField("Name", "")
	tt.NotNil(sliceBuilder.Interface())
}

func TestStructBuilderMergeEdgeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

	builder := ztype.NewStruct()

	err := builder.Merge(123)
	tt.EqualTrue(err != nil)

	type Sample struct {
		Name string
		Age  int
	}

	err = builder.Merge(Sample{Name: "test", Age: 25})
	tt.NoError(err)

	builder2 := ztype.NewStruct()
	err = builder2.Merge(
		Sample{Name: "first", Age: 20},
		Sample{Name: "second", Age: 30},
	)
	tt.NoError(err)
}

func TestCacheParseStructFieldsEdgeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

	type Simple struct {
		Field1 string `z:"field1"`
		Field2 int    `z:"field2"`
	}

	s := Simple{
		Field1: "value1",
		Field2: 42,
	}

	m := ztype.ToMap(s)
	tt.Equal("value1", m.Get("field1").String())
	tt.Equal(42, m.Get("field2").Int())

	type Empty struct{}
	e := Empty{}
	m2 := ztype.ToMap(e)
	tt.NotNil(m2)
}
