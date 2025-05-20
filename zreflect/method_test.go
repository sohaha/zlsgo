package zreflect

import (
	"reflect"
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestGetAllMethod(t *testing.T) {
	tt := zlsgo.NewTest(t)
	i := 0
	GetAllMethod(Demo, func(numMethod int, m reflect.Method) error {
		t.Log(numMethod, m.Name)
		i++
		switch numMethod {
		case 0:
			tt.Equal("Text", m.Name)
		}
		return nil
	})
	tt.Equal(1, i)
}

func TestRunAssignMethod(t *testing.T) {
	tt := zlsgo.NewTest(t)
	i := 0
	RunAssignMethod(Demo, func(methodName string) bool {
		t.Log("methodName:", methodName)
		i++
		return false
	})
	tt.Equal(1, i)
}
