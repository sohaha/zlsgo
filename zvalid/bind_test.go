package zvalid

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestVar(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var str string
	err := Var(&str, Text("is var").RemoveSpace())
	tt.EqualNil(err)
	tt.Equal("isvar", str)

	var i int
	err = Var(&i, Text("is var").RemoveSpace())
	tt.Equal(true, err != nil)
	tt.Equal(0, i)
	err = Var(&i, Text("99").RemoveSpace())
	tt.EqualNil(err)
	tt.Equal(99, i)

	var sts []string
	err = Var(&sts, Text("1,2,3,go").Separator(","))
	tt.EqualNil(err)
	tt.Equal([]string{"1", "2", "3", "go"}, sts)

	var data struct {
		Name string
	}

	err = Batch(
		BatchVar(&data.Name, Text("yes name")),
	)
	tt.EqualNil(err)
	tt.Equal("yes name", data.Name)

}
