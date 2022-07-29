package zvalid

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zjson"
)

func TestValids(t *testing.T) {
	tt := zlsgo.NewTest(t)

	j := zjson.Parse(`{"name":"zls","age":18}`)
	rules := map[string]Engine{
		"name": New().MinLength(3),
		"age":  New().MinInt(18),
	}

	err := JSON(j, rules)
	tt.NoError(err)

	j = zjson.Parse(`{"age":8}`)
	err = JSON(j, rules)
	t.Log(err)
	tt.EqualTrue(err != nil)

	j = zjson.Parse(`{"name":8}`)
	err = JSON(j, rules)
	t.Log(err)
	tt.EqualTrue(err != nil)

	j = zjson.Parse(``)
	err = JSON(j, map[string]Engine{
		"password": New().Required("密码不能为空"),
	})
	t.Log(err)
	tt.EqualTrue(err != nil)

	j = zjson.Parse(`{"password":"123456"}`)
	err = JSON(j, map[string]Engine{
		"password": New().Required("密码不能为空").StrongPassword("密码必须是强密码"),
	})
	t.Log(err)
	tt.EqualTrue(err != nil)

	j = zjson.Parse(`{"password":"123456Abc."}`)
	err = JSON(j, map[string]Engine{
		"password": New().Required("密码不能为空").StrongPassword("密码必须是强密码"),
	})
	tt.NoError(err)

}
