package znet

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zvalid"
)

func TestBindValid(t *testing.T) {
	tt := zlsgo.NewTest(t)
	r := newServer()
	_ = newRequest(r, "POST", []string{"/TestBindValid", `{"id":666,"Pid":100,"name":"HelloWorld","g":{"Info":"基础"},"ids":[{"id":1,"Name":"用户1","g":{"Info":"详情","p":[{"id":1},{"id":2}]}}]}`, mimeJSON}, "/TestBindValid", func(c *Context) {
		var s SS
		r := c.ValidRule().Required()
		err := c.BindValid(&s, map[string]zvalid.Engine{
			"id": r.IsNumber().Customize(func(rawValue string, err error) (newValue string, newErr error) {
				newValue = "1999"
				return
			}),
			"name": r.CamelCaseToSnakeCase(),
		})
		t.Logf("%+v", s)
		t.Log(err)
		tt.EqualNil(err)
		tt.Equal("hello_world", s.Name)
		tt.Equal(1, s.IDs[0].Gg.P[0].ID)
	})

}

func TestContextValid(t *testing.T) {
	tt := zlsgo.NewTest(t)
	r := newServer()
	w := newRequest(r, "POST", []string{
		"/TestContext_Valid?body2=b2",
		"body=1 &b3=999", "application/x-www-form-urlencoded",
	}, "/TestContext_Valid", func(c *Context) {
		var rbool bool
		rule := c.ValidRule().HasNumber()

		v, err := c.Valid(rule, "body", "内容").MinLength(5).IsNumber().String()

		tt.Equal(true, err != nil)
		t.Log(v, err)

		v, err = c.Valid(rule, "body2", "内容2").Required().MinLength(5).String()
		tt.Equal(true, err != nil)
		t.Log(v, err)

		_, _ = c.ValidParam(rule, "body2").Required().String()

		err = c.ValidQuery(rule, "body2", "内容2-2").Required().Error()
		tt.Equal(true, err == nil)
		t.Log(err)

		rbool, err = c.ValidForm(rule, "body", "内容1-2").Required().Trim().Bool()
		tt.Equal(true, err == nil)
		tt.Equal(true, rbool)
		t.Log(rbool, err)

		c.String(200, expected)
	})
	tt.Equal(200, w.Code)
	tt.Equal(expected, w.Body.String())
}

func TestContextBatchValid(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	r := newServer()
	w := newRequest(r, "POST", []string{
		"/TestContext_BatchValid?is=json",
		`{"title":"这是json"}`, "application/json",
	}, "/TestContext_BatchValid", func(c *Context) {
		c.WithValue("varName", "666")
		tt.Log("==前置")
		c.Next()
		tt.Log("==后置")
		// var log bytes.Buffer
		// rsp := io.MultiWriter(c.Writer, &log)
		// tt.Log(rsp)
		// tt.Log(log.String())
		// tt.Log(fmt.Sprintf("%#v", c.Request.Response))
		// tt.Log(fmt.Sprintf("%#v", c.Writer))
		// tt.Log(c.Request)
		// tt.Log(c.Code)
	}, func(c *Context) {
		tt.Log("--前置2")
		c.Abort()
		c.Next()
	}, func(c *Context) {
		// rule := zvalid.New().HasNumber()
		raw, _ := c.GetDataRaw()
		tt.Log(raw)
		tt.Log(c.Value("varName"))
		tt.Log(c.Value("varName2"))
		tt.Log(c.Value("varName3", []string{"a", "varName3"}))

		json, err := c.GetJSONs()
		tt.Log(json, err)
		tt.Log(json.String())
		t.Equal(c.GetJSON("title").String(), json.Get("title").String())

		tt.Log(c.GetJSON("title"))

		v, err := c.ValidJSON(c.ValidRule(), "title", "内容2-2").Required().String()
		t.Equal(true, err == nil)
		tt.Log(v, err)

		c.String(200, "")
	})
	t.Equal(200, w.Code)
}
