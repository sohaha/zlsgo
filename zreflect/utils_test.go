package zreflect

import (
	"reflect"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
)

type (
	Child2 struct {
		P DemoChildSt
	}

	DemoChildSt struct {
		ChildName int
	}

	DemoSt struct {
		Date2  time.Time
		Child3 *DemoChildSt
		Remark string `json:"remark"`
		note   string
		Any    interface{}
		any    interface{}
		Name   string `json:"username"`
		Slice  [][]string
		Hobby  []string
		Child  struct {
			Title       string `json:"child_user_title"`
			DemoChild2  Child2 `json:"demo_child_2"`
			IsChildName bool
		} `json:"child"`
		Year   float64
		Child2 Child2
		child4 DemoChildSt
		Age    uint
		Lovely bool
		pri    string
	}
	TestSt struct {
		Name string
		I    int `z:"iii"`
		Note int `json:"note,omitempty"`
	}
)

func (d DemoSt) Text() string {
	return d.Name + ":" + d.Remark
}

func TestNonzero(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Equal(true, Nonzero(ValueOf(true)))
	tt.Equal(false, Nonzero(ValueOf(0)))
	tt.Equal(true, Nonzero(ValueOf(1)))
	tt.Equal(true, Nonzero(ValueOf("0")))
	tt.Equal(true, Nonzero(ValueOf("1")))
	tt.Equal(true, Nonzero(ValueOf(1.1)))
	var s []string
	tt.Equal(false, Nonzero(ValueOf(s)))
	tt.Equal(true, Nonzero(ValueOf([]string{})))
	tt.Equal(false, Nonzero(ValueOf([...]string{})))
	tt.Equal(true, Nonzero(ValueOf(map[string]string{})))
	tt.Equal(true, Nonzero(ValueOf(map[string]string{"a": "b"})))
}

func TestCan(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Equal(true, IsLabel(TypeOf(Demo)))

	tt.Equal(false, CanExpand(TypeOf(1)))
	tt.Equal(true, CanExpand(TypeOf([]string{})))

	tt.Equal(true, CanInline(TypeOf(map[string]string{})))
	tt.Equal(true, CanInline(TypeOf([]string{})))
	tt.Equal(true, CanInline(TypeOf("1")))
	tt.Equal(false, CanInline(TypeOf(&Demo)))
	tt.Equal(false, CanInline(TypeOf(Demo)))
	tt.Equal(false, CanInline(TypeOf(func() {})))
}

func TestGetAbbrKind(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Equal(reflect.Int, GetAbbrKind(ValueOf(1)))
	tt.Equal(reflect.Uint, GetAbbrKind(ValueOf(uint64(1))))
	tt.Equal(reflect.Float64, GetAbbrKind(ValueOf(float32(1))))
	tt.Equal(reflect.Struct, GetAbbrKind(ValueOf(Demo)))
}
