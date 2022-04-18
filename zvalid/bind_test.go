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

	var iu uint
	err = Var(&iu, Text("99").RemoveSpace())
	tt.EqualNil(err)
	tt.Equal(uint(99), iu)

	var f32 float32
	val := Text("99.0")
	err = Var(&f32, val)
	tt.EqualNil(err)
	tt.Equal(float32(99), f32)

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

func TestVarDefault(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var email string
	err := Var(&email, Text("email").IsMail())
	t.Log(email, err)
	tt.EqualExit(email, "")
	tt.EqualTrue(err != nil)

	err = Var(&email, Text("email").IsMail().Default(666))
	t.Log(email, err)
	tt.EqualExit(email, "")
	tt.EqualTrue(err != nil)

	err = Var(&email, Text("email").Silent().IsMail().Default("qq@qq.com"))
	t.Log(email, err)
	tt.EqualExit("qq@qq.com", email)
	tt.ErrorNil(err)

	err = Var(&email, Text("email").IsMail().Default("qq@qq.com"))
	t.Log(email, err)
	tt.EqualExit(email, "qq@qq.com")
	tt.ErrorNil(err)

	var nu int
	err = Var(&nu, Text("Number").IsNumber().Default(123))
	t.Log(nu, err)
	tt.ErrorNil(err)
	tt.EqualExit(nu, 123)

	var b bool
	err = Var(&b, Text("true").IsBool().Default(false))
	t.Log(b, err)
	tt.EqualTrue(err == nil)
	tt.EqualExit(b, true)

	var i uint
	err = Var(&i, Text("true").IsNumber().Default(uint(123)))
	t.Log(b, err)
	tt.ErrorNil(err)
	tt.EqualExit(uint(123), i)

	var f float32
	err = Var(&f, Text("true").IsNumber().Default(float32(123)))
	t.Log(b, err)
	tt.ErrorNil(err)
	tt.EqualExit(float32(123), f)
}
