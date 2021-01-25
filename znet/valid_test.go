package znet

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestContextValid(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	r := newServer()
	w := newRequest(r, "POST", []string{
		"/TestContext_Valid?body2=b2",
		"body=1 &b3=999", "application/x-www-form-urlencoded",
	}, "/TestContext_Valid", func(c *Context) {
		var rbool bool
		rule := c.ValidRule().HasNumber()

		v, err := c.Valid(rule, "body", "内容").MinLength(5).IsNumber().String()

		t.Equal(true, err != nil)
		tt.Log(v, err)

		v, err = c.Valid(rule, "body2", "内容2").Required().MinLength(5).String()
		t.Equal(true, err != nil)
		tt.Log(v, err)

		_, _ = c.ValidParam(rule, "body2").Required().String()

		err = c.ValidQuery(rule, "body2", "内容2-2").Required().Error()
		t.Equal(true, err == nil)
		tt.Log(err)

		rbool, err = c.ValidForm(rule, "body", "内容1-2").Required().Trim().Bool()
		t.Equal(true, err == nil)
		t.Equal(true, rbool)
		tt.Log(rbool, err)

		c.String(200, expected)
	})
	t.Equal(200, w.Code)
	t.Equal(expected, w.Body.String())
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
