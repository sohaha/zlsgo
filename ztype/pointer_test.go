//go:build go1.18
// +build go1.18

package ztype_test

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/ztype"
)

func TestNewPointer(t *testing.T) {
	tt := zlsgo.NewTest(t)

	b := ztype.ToPointer(true)
	tt.EqualExit(true, *b)

	b2 := ztype.ToPointer(false)
	tt.EqualExit(false, *b2)

	i := ztype.ToPointer(1)
	tt.EqualExit(1, *i)
}
