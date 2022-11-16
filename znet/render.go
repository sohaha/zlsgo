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
	"strings"

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
		ContentDate []byte
		FileExist   bool
	}
	renderHTML struct {
		Template    *template.Template
		Data        interface{}
		ContentDate []byte
		FuncMap     template.FuncMap
		Templates   []string
	}
	// ApiData unified return api format
	ApiData struct {
		Data interface{} `json:"data"`
		Msg  string      `json:"msg,omitempty"`
		Code int32       `json:"code" example:"200"`
	}
	// Data map string
	Data     map[string]interface{}
	PrevData struct {
		Code    *zutil.Int32
		Type    string
		Content []byte
	}
)

var (
	// ContentTypePlain text
	ContentTypePlain = "text/plain; charset=utf-8"
	// ContentTypeHTML html
	ContentTypeHTML = "text/html; charset=utf-8"
	// ContentTypeJSON json
	ContentTypeJSON = "application/json; charset=utf-8"
)

func (c *Context) renderProcessing(code int32, r render) {
	if c.stopHandle.Load() && c.prevData.Code.Load() != 0 {
		return
	}
	c.prevData.Code.Store(code)
	c.mu.Lock()
	c.render = r
	c.mu.Unlock()
}

func (r *renderByte) Content(c *Context) []byte {
	if !c.hasContentType() {
		c.SetContentType(ContentTypePlain)
	}
	return r.Data
}

func (r *renderString) Content(c *Context) []byte {
	if r.ContentDate != nil {
		return r.ContentDate
	}
	if !c.hasContentType() {
		c.SetContentType(ContentTypePlain)
	}
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
	if len(r.Templates) > 0 {
		var (
			buf bytes.Buffer
			err error
			t   *template.Template
		)
		tpl := c.Engine.template
		if tpl != nil {
			t = tpl.Get(c.Engine.IsDebug())
			if t != nil && len(r.FuncMap) == 0 {
				name := r.Templates[0]
				err = t.ExecuteTemplate(&buf, name, r.Data)
				if err == nil {
					r.ContentDate = buf.Bytes()
					return r.ContentDate
				}
				if !strings.Contains(err.Error(), " is undefined") {
					panic(err)
				}
			}
		}
		if t, err = templateParse(r.Templates, r.FuncMap); err == nil {
			err = t.Execute(&buf, r.Data)
		}
		if err != nil {
			panic(err)
		}
		r.ContentDate = buf.Bytes()
	} else {
		r.ContentDate = zstring.String2Bytes(fmt.Sprint(r.Data))
	}
	return r.ContentDate
}

func (c *Context) Byte(code int32, value []byte) {
	c.renderProcessing(code, &renderByte{Data: value})
}

func (c *Context) String(code int32, format string, values ...interface{}) {
	c.renderProcessing(code, &renderString{Format: format, Data: values})
}

// Deprecated: You can directly modify the return value of PrevContent()
func (c *Context) SetContent(data *PrevData) {
	c.mu.Lock()
	c.prevData = data
	c.mu.Unlock()
}

func (c *Context) File(path string) {
	path = zfile.RealPath(path)
	f, err := os.Stat(path)
	fileExist := err == nil
	var code int32
	if fileExist {
		code = http.StatusOK
	} else {
		code = http.StatusNotFound
	}
	if fileExist {
		c.SetHeader("Last-Modified", f.ModTime().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	}
	c.renderProcessing(code, &renderFile{Data: path, FileExist: fileExist})
}

func (c *Context) JSON(code int32, values interface{}) {
	c.renderProcessing(code, &renderJSON{Data: values})
}

// ApiJSON ApiJSON
func (c *Context) ApiJSON(code int32, msg string, data interface{}) {
	c.renderProcessing(http.StatusOK, &renderJSON{Data: ApiData{Code: code, Data: data,
		Msg: msg}})
}

// HTML export html
func (c *Context) HTML(code int32, html string) {
	c.renderProcessing(code, &renderHTML{
		Data: html,
	})
}

// Template export tpl
func (c *Context) Template(code int32, name string, data interface{}, funcMap ...map[string]interface{}) {
	var fn template.FuncMap
	if len(funcMap) > 0 {
		fn = funcMap[0]
	}
	c.renderProcessing(code, &renderHTML{
		Templates: []string{name},
		Data:      data,
		FuncMap:   fn,
	})
}
func (c *Context) Templates(code int32, templates []string, data interface{}, funcMap ...map[string]interface{}) {
	var fn template.FuncMap
	if len(funcMap) > 0 {
		fn = funcMap[0]
	}
	c.renderProcessing(code, &renderHTML{
		Templates: templates,
		Data:      data,
		FuncMap:   fn,
	})
}

// Abort stop executing subsequent handlers
func (c *Context) Abort(code ...int32) {
	if c.stopHandle.Load() {
		return
	}
	c.stopHandle.Store(true)
	if len(code) > 0 {
		c.prevData.Code.Store(code[0])
	}
}

// Redirect Redirect
func (c *Context) Redirect(link string, statusCode ...int32) {
	c.Writer.Header().Set("Location", c.CompletionLink(link))
	var code int32
	if len(statusCode) > 0 {
		code = statusCode[0]
	} else {
		code = http.StatusFound
	}
	c.SetStatus(code)
}

func (c *Context) SetStatus(code int32) *Context {
	c.prevData.Code.Store(code)
	return c
}

func (c *Context) SetContentType(contentType string) *Context {
	c.SetHeader("Content-Type", contentType)
	return c
}

func (c *Context) hasContentType() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if _, ok := c.header["Content-Type"]; ok {
		return true
	}
	return false
}

// PrevContent current output content
func (c *Context) PrevContent() *PrevData {
	if c.render == nil {
		return c.prevData
	}
	c.prevData.Content = c.render.Content(c)
	ctype, hasType := c.header["Content-Type"]
	if hasType {
		c.prevData.Type = ctype[0]
	}
	c.mu.Lock()
	c.render = nil
	c.mu.Unlock()
	return c.prevData
}

func (t *tpl) Get(debug bool) *template.Template {
	if !debug || t.pattern == "" {
		return t.tpl
	}
	tpl, _ := template.New("").Funcs(t.templateFuncMap).ParseGlob(t.pattern)
	return tpl
}
