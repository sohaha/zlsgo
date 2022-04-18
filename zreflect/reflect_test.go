package zreflect_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zreflect"
)

type (
	DemoSt struct {
		Name  string `json:"username"`
		Age   uint
		Child struct {
			IsChildName bool
			Title       string `json:"child_user_title"`
			DemoChild2  Child2 `json:"demo_child_2"`
		} `json:"child"`
		Hobby  []string
		note   string
		Lovely bool
		Year   float64
		Date2  time.Time
		Remark string `json:"remark"`
		Child2 Child2
		Child3 *DemoChildSt
		child4 DemoChildSt
		Slice  [][]string
	}
	TestSt struct {
		Name string
		I    int `z:"iii"`
		Note int `json:"note,omitempty"`
	}
)

func (d DemoSt) name() {
	fmt.Println("name", d.Name)
}

func (d *DemoSt) name2() {
	fmt.Println("name2", d.Name)
}

func (d DemoSt) Name3() {
	fmt.Println("name3", d.Name)
}

func (d *DemoSt) Name4() {
	fmt.Println("name4", d.Name)
}

type Child2 struct {
	P DemoChildSt
}

type DemoChildSt struct {
	ChildName int
}

var data = map[string]interface{}{
	"username": "The is Username",
	"note":     "The is note",
	"Age":      999,
	"Lovely":   true,
	"Date2":    "2022-04-03 15:34:42",
	"Child2": map[string]map[string]int{
		"P": {
			"ChildName": 121,
		},
	},
	"Hobby": []string{"Go", "😴"},
	"child": map[string]interface{}{
		"IsChildName":      true,
		"child_user_title": "child title",
		"demo_child_2": map[string]interface{}{
			"P": map[string]interface{}{"ChildName": "P_ChildName"},
		},
	},
	"Slice": [][]string{{"1", "2"}, {"3", "4"}},
}

func TestRegister(t *testing.T) {
	tt := zlsgo.NewTest(t)
	d := DemoSt{}
	err := zreflect.Register(d)
	t.Log(err)
	tt.EqualNil(err)

	var i int
	err = zreflect.Register(&i)
	t.Log(err)
	tt.EqualTrue(err != nil)

	var n struct{ Name string }
	err = zreflect.Register(&n)
	t.Log(err)
	tt.EqualNil(err)

	type stn struct{ Name string }
	var vstn stn
	err = zreflect.Register(&vstn)
	t.Log(err)
	tt.EqualNil(err)
}

func TestNewVal(t *testing.T) {
	tt := zlsgo.NewTest(t)
	var demo DemoSt

	v, err := zreflect.ValueOf(&demo)
	tt.ErrorNil(err)

	val, err := zreflect.NewVal(v)
	tt.ErrorNil(err)

	t.Log(val.Name())

	err = val.ForEachVal(func(parent []string, index int, tag string, field reflect.StructField, val reflect.Value) error {
		t.Log(parent, index, tag, field, val)
		if tag == "username" {
			val.SetString("zlsgo")
		} else if tag == "IsChildName" {
			val.SetBool(true)
		}
		return nil
	})
	tt.ErrorNil(err)
	tt.Equal("zlsgo", demo.Name)
	tt.Equal(true, demo.Child.IsChildName)

	v, err = zreflect.ValueOf(demo)
	tt.EqualTrue(err != nil)

	val, err = zreflect.NewVal(v)
	tt.EqualTrue(err != nil)
}

func TestNewTyp(t *testing.T) {
	tt := zlsgo.NewTest(t)
	var demo DemoSt

	tp := zreflect.TypeOf(&demo)

	typ, err := zreflect.NewTyp(tp)
	tt.EqualNil(err)

	err = typ.ForEachVal(func(parent []string, index int, tag string, field reflect.StructField, val reflect.Value) error {
		t.Log(parent, index, tag, field, val)
		return nil
	})
	tt.EqualTrue(err != nil)

	totalSkipStruct := 0
	err = typ.ForEach(func(parent []string, index int, tag string, field reflect.StructField) error {
		t.Log(parent, index, tag, field)
		tt.EqualTrue(len(parent) == 0)
		totalSkipStruct++
		return zreflect.ErrSkipStruct
	})
	tt.EqualNil(err)

	total := 0
	err = typ.ForEach(func(parent []string, index int, tag string, field reflect.StructField) error {
		t.Log(parent, index, tag, field)
		total++
		return nil
	})
	tt.EqualNil(err)
	tt.EqualTrue(total > totalSkipStruct)
}

func initBenchmarkMaps(demo interface{}) (*zreflect.Typer, map[string]interface{}) {
	v, _ := zreflect.ValueOf(demo)
	val, _ := zreflect.NewVal(v)
	data := map[string]interface{}{
		"username": "The is Username",
		"note":     "The is note",
		"Age":      999,
	}
	return val, data
}

func BenchmarkMapTypStruct1(b *testing.B) {
	var demo DemoSt
	t, from := initBenchmarkMaps(&demo)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := zreflect.MapTypStruct(from, t)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkMapTypStruct2(b *testing.B) {
	var demo DemoSt
	t, from := initBenchmarkMaps(&demo)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := zreflect.MapTypStruct(from, t)
		if err != nil {
			b.Error(err)
		}
	}
}
