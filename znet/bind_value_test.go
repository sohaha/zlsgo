package znet

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestContext_Bind(t *testing.T) {
	tt := zlsgo.NewTest(t)
	r := newServer()
	_ = newRequest(r, "POST", []string{"/TestContext_BindJSON", `{"id":666,"Pid":100,"name":"名字","g":{"Info":"基础"},"ids":[{"id":1,"Name":"用户1","g":{"Info":"详情","p":[{"id":1},{"id":2}]}}]}`, mimeJSON}, "/TestContext_BindJSON", func(c *Context) {
		var s SS
		err := c.Bind(&s)
		t.Logf("%+v", s)
		tt.EqualNil(err)
		tt.Equal("名字", s.Name)
		tt.Equal(1, s.IDs[0].Gg.P[0].ID)
	})
}

func TestContext_BindJSON(t *testing.T) {
	tt := zlsgo.NewTest(t)
	r := newServer()
	_ = newRequest(r, "POST", []string{"/TestContext_BindJSON", `{"id":666,"Pid":100,"name":"名字","g":{"Info":"基础"},"ids":[{"id":1,"Name":"用户1","g":{"Info":"详情"}}]}`, mimeJSON}, "/TestContext_BindJSON", func(c *Context) {
		var s SS
		err := c.BindJSON(&s)
		tt.Log(s)
		tt.EqualNil(err)
		tt.Equal("名字", s.Name)
	})
}

func TestContext_BindQuery(t *testing.T) {
	tt := zlsgo.NewTest(t)
	r := newServer()
	_ = newRequest(r, "Get", "/TestContext_BindQuery?id=666&&t=1&t=2&t2=1&t2=2&g[Info]=基础&name=_name&ids[1][id]=123&ids[0][Name]=ids_0_name&p[n]=is pn&p[Key]=1.234", "/TestContext_BindQuery", func(c *Context) {
		var s SS
		err := c.BindQuery(&s)
		tt.EqualNil(err)
		tt.Equal("_name", s.Name)
		tt.Equal("is pn", s.Property.Name)
		tt.Equal(1, s.To2)
		tt.Equal([]string{"1", "2"}, s.To)
		t.Logf("%+v\n", s)
	})
}

func TestContext_BindForm(t *testing.T) {
	tt := zlsgo.NewTest(t)
	r := newServer()
	_ = newRequest(r, "POST", []string{"/TestContext_BindForm", `id=666&&t=1&t=2&t2=1&t2=2&g[Info]=基础&name=_name&ids[1][id]=123&ids[0][Name]=ids_0_name&p[n]=is pn&p[Key]=1.234`, mimePOSTForm}, "/TestContext_BindForm", func(c *Context) {
		var s SS
		err := c.BindForm(&s)
		tt.EqualNil(err)
		tt.Equal("_name", s.Name)
		tt.Equal("is pn", s.Property.Name)
		tt.Equal(1, s.To2)
		tt.Equal([]string{"1", "2"}, s.To)
		t.Logf("%+v\n", s)
	})
}
