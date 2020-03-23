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
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zutil"
)

type (
	render interface {
		Content(c *Context) (content []byte)
	}
	renderByte struct {
		Data        []byte
		Type        string
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
		Data        string
		FileExist   bool
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
	Data     map[string]interface{}
	PrevData struct {
		Code    int
		Type    string
		Content []byte
	}
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

func (r *renderByte) Content(c *Context) []byte {
	c.SetContentType(zutil.IfVal(r.Type != "", r.Type, ContentTypePlain).(string))
	return r.Data
}

func (r *renderString) Content(c *Context) []byte {
	if r.ContentDate != nil {
		return r.ContentDate
	}
	c.SetContentType(ContentTypePlain)
	if len(r.Data) > 0 {
		r.ContentDate = zstring.String2Bytes(fmt.Sprintf(r.Format, r.Data...))
	} else {
		r.ContentDate = zstring.String2Bytes(r.Format)
	}
	return r.ContentDate
}

func (r *renderJSON) Content(c *Context) []byte {
	if r.ContentDate != nil {
		return r.ContentDate
	}
	c.SetContentType(ContentTypeJSON)
	r.ContentDate, _ = json.Marshal(r.Data)
	return r.ContentDate
}

func (r *renderFile) Content(c *Context) []byte {
	if !r.FileExist {
		return []byte{}
	}

	if r.ContentDate != nil {
		return r.ContentDate
	}

	fType := mime.TypeByExtension(filepath.Ext(r.Data))
	c.SetContentType(fType)
	r.ContentDate, _ = ioutil.ReadFile(r.Data)
	return r.ContentDate
}

func (r *renderHTML) Content(c *Context) []byte {
	if r.ContentDate != nil {
		return r.ContentDate
	}
	c.SetContentType(ContentTypeHTML)
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

func (c *Context) Byte(code int, value []byte) {
	c.render(code, &renderByte{Data: value})
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.render(code, &renderString{Format: format, Data: values})
}

func (c *Context) SetContent(data *PrevData) {
	c.render(data.Code, &renderByte{Data: data.Content, Type: data.Type})
}

func (c *Context) File(path string) {
	fileExist := zfile.FileExist(path)
	code := zutil.IfVal(fileExist, 200, 404).(int)
	if fileExist {
		if f, err := os.Stat(path); err == nil {
			c.SetHeader("Last-Modified", f.ModTime().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT"))
		}
	}
	c.render(code, &renderFile{Data: path, FileExist: fileExist})
}

func (c *Context) JSON(code int, values interface{}) {
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

func (c *Context) PrevContent() *PrevData {
	c.Info.Mutex.RLock()
	r := c.Info.render
	code := c.Info.Code
	ctype, hasType := c.Info.heades["Content-Type"]
	c.Info.Mutex.RUnlock()
	if !hasType {
		ctype = ContentTypePlain
	}
	var content []byte
	if r != nil {
		content = r.Content(c)
	}
	return &PrevData{
		Code:    code,
		Type:    ctype,
		Content: content,
	}
}
