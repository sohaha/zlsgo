package zutil

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

type TestSt struct {
}

func (*TestSt) RunTest(t *testing.T) {
	t.Log("RunTest")
}

func TestRunAllMethod(t *testing.T) {
	tt := zlsgo.NewTest(t)
	err := RunAllMethod(&TestSt{}, t)
	tt.EqualNil(err)
}
