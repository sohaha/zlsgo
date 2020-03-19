/*
 * @Author: seekwe
 * @Date:   2019-05-23 19:16:32
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-05-29 15:40:30
 */

package znet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zutil"
)

type (
	render interface {
		Render(*Context, int) error
		Content() (content []byte)
	}
	renderByte struct {
		Data        []byte
		ContentDate []byte
	}
	renderString struct {
		Format      string
		Data        []interface{}
		ContentDate []byte
	}
	renderJSON struct {
		Data        interface{}
		ContentDate []byte
	}
	renderFile struct {
		Status      interface{}
		Data        interface{}
		ContentDate []byte
	}
	renderHTML struct {
		Template    *template.Template
		Name        string
		Data        interface{}
		ContentDate []byte
	}
	Api struct {
		Code int         `json:"code" example:"200"`
		Msg  string      `json:"msg"`
		Data interface{} `json:"data"`
	}
	Data map[string]interface{}
)

var (
	ContentTypePlain = "text/plain; charset=utf-8"
	ContentTypeHTML  = "text/html; charset=utf-8"
	ContentTypeJSON  = "application/json; charset=utf-8"
)

func (c *Context) render(code int, r render) {
	c.Info.Mutex.Lock()
	c.Info.Code = code
	c.Info.render = r
	c.Info.StopHandle = true
	c.Info.Mutex.Unlock()
}

func (r *renderByte) Content() []byte {
	return r.Data
}

func (r *renderByte) Render(c *Context, code int) (err error) {
	w := c.Writer
	c.SetStatus(code)
	_, err = w.Write(r.Data)
	return
}

func (r *renderString) Content() []byte {
	if r.ContentDate != nil {
		return r.ContentDate
	}
	if len(r.Data) > 0 {
		r.ContentDate = zstring.String2Bytes(fmt.Sprintf(r.Format, r.Data...))
	} else {
		r.ContentDate = zstring.String2Bytes(r.Format)
	}
	return r.ContentDate
}

func (r *renderString) Render(c *Context, code int) (err error) {
	w := c.Writer
	c.SetStatus(code)
	_, err = w.Write(r.Content())
	return
}

func (r *renderJSON) Content() []byte {
	if r.ContentDate != nil {
		return r.ContentDate
	}
	r.ContentDate, _ = json.Marshal(r.Data)
	return r.ContentDate
}

func (r *renderFile) Content() []byte {
	if r.ContentDate != nil {
		return r.ContentDate
	}
	r.ContentDate = zstring.String2Bytes(r.Data.(string))
	return r.ContentDate
}

func (r *renderFile) Render(c *Context, _ int) error {
	http.ServeFile(c.Writer, c.Request, r.Data.(string))
	return nil
}

func (r *renderJSON) Render(c *Context, code int) error {
	w := c.Writer
	c.SetStatus(code)
	_, _ = w.Write(r.Content())
	return nil
}

func (r *renderHTML) Content() []byte {
	if r.ContentDate != nil {
		return r.ContentDate
	}
	if r.Name != "" {
		var t *template.Template
		var err error
		t, err = templateParse(r.Name)
		if err != nil {
			return r.ContentDate
		}
		var buf bytes.Buffer
		err = t.Execute(&buf, r.Data)
		if err != nil {
			return r.ContentDate
		}
		r.ContentDate = buf.Bytes()
	} else {
		r.ContentDate = zstring.String2Bytes(fmt.Sprint(r.Data))
	}
	return r.ContentDate
}

func (r *renderHTML) Render(c *Context, code int) (err error) {
	w := c.Writer
	c.SetStatus(code)
	_, _ = w.Write(r.Content())
	return
}

func (c *Context) Byte(code int, value []byte) {
	c.SetContentType(ContentTypePlain)
	c.render(code, &renderByte{Data: value})
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetContentType(ContentTypePlain)
	c.render(code, &renderString{Format: format, Data: values})
}

func (c *Context) File(filepath string) {
	c.render(zutil.IfVal(zfile.FileExist(filepath), 200, 404).(int), &renderFile{Data: filepath})
}

func (c *Context) JSON(code int, values interface{}) {
	c.SetContentType(ContentTypeJSON)
	c.render(code, &renderJSON{Data: values})
}

// ResJSON ResJSON
func (c *Context) ResJSON(code int, msg string, data interface{}) {
	httpState := code
	if code < 300 && code >= 200 {
		httpState = http.StatusOK
	}
	c.render(httpState, &renderJSON{Data: Api{Code: code, Data: data, Msg: msg}})
}

// HTML export html
func (c *Context) HTML(code int, html string) {
	c.SetContentType(ContentTypeHTML)
	c.render(code, &renderHTML{
		Name: "",
		Data: html,
	})
}

// Template export template
func (c *Context) Template(code int, name string, data ...interface{}) {
	var _data interface{}
	if len(data) > 0 {
		_data = data[0]
	}
	c.render(code, &renderHTML{
		Name: name,
		Data: _data,
	})
}

// Abort Abort
func (c *Context) Abort(code ...int) {
	c.Info.Mutex.Lock()
	c.Info.StopHandle = true
	c.Info.Mutex.Unlock()
	if len(code) > 0 {
		c.SetStatus(code[0])
	}
}

// Redirect Redirect
func (c *Context) Redirect(link string, statusCode ...int) {
	c.Writer.Header().Set("Location", c.CompletionLink(link))
	code := http.StatusFound
	if len(statusCode) > 0 {
		code = statusCode[0]
	}
	c.SetStatus(code)
}

func (c *Context) SetStatus(code int) *Context {
	c.Info.Mutex.Lock()
	c.Info.Code = code
	c.Info.Mutex.Unlock()
	return c
}

func (c *Context) SetContentType(contentType string) *Context {
	c.SetHeader("Content-Type", contentType)
	return c
}

func (c *Context) GetContentType() string {
	c.Info.Mutex.RLock()
	value, _ := c.Info.heades["Content-Type"]
	c.Info.Mutex.RUnlock()
	return value
}

// PrevStatus current http status code
func (c *Context) PrevStatus() (code int) {
	c.Info.Mutex.RLock()
	code = c.Info.Code
	c.Info.Mutex.RUnlock()
	return
}
