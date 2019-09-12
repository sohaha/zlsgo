package ztype

import (
	"github.com/sohaha/zlsgo"
	"testing"
)

type T1 struct {
	V1 int
}
type T2 struct {
	V2 int
}

type VV struct {
	V0 interface{}
	T1
	T2
	T4 string `z:"ignore"`
}

func TestStructGetFields(T *testing.T) {
	t := zlsgo.NewTest(T)
	e := Struct()
	var n = new(VV)
	n.T4 = "t4"
	res := e.GetStructFields(n)
	for _, item := range res {
		switch ty := item.(type) {
		case *int:
			t.Log("int")
			*ty = 666
		case *string:
			t.Log("string")
			*ty = "t4-666"
		default:
			t.Log(GetType(item))
		}
	}
	t.Equal(666, n.V2)
	t.Log(n)
}

func TestStruct(T *testing.T) {
	t := zlsgo.NewTest(T)
	e := Struct()
	vars := &VV{
		V0: "interface",
		T1: T1{
			V1: 1,
		},
		T2: T2{
			V2: 2,
		},
		T4: "t4",
	}
	vars2 := new(VV)
	vars2.V2 = 22
	vars2.V0 = 20
	vars2.T4 = "is too t4"
	varsM := []*VV{
		vars,
		vars2,
	}
	res := e.ToMap(varsM)
	t.Log(res)
}
