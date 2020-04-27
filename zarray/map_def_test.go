package zarray

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestOptionData(t *testing.T) {
	tt := zlsgo.NewTest(t)
	option := &DefData{
		"string":     nil,
		"int":        nil,
		"bool":       nil,
		"float64":    nil,
		"funcSingle": nil,
	}
	tt.Equal("ss", option.String("string", "ss"))
	tt.Equal(11, option.Int("int", 11))
	tt.Equal(true, option.Bool("bool", true))
	tt.Equal(1.2, option.Float64("float64", 1.2))
	option.FuncSingle("funcSingle", func() {})
	option = &DefData{
		"string":     "s",
		"int":        1,
		"bool":       true,
		"float64":    1.2,
		"funcSingle": func() {},
	}
	tt.Equal("s", option.String("string", "ss"))
	tt.Equal(1, option.Int("int", 11))
	tt.Equal(true, option.Bool("bool", true))
	tt.Equal(1.2, option.Float64("float64", 1.1))
	option.FuncSingle("funcSingle", func() {})
}
