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
		FileExist   bool
		ContentDate []byte
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
		Code int         `json:"code" example:"200"`
		Msg  string      `json:"msg"`
		Data interface{} `json:"data"`
	}
	// Data map string
	Data     map[string]interface{}
	PrevData struct {
		Code    int
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

func (c *Context) renderProcessing(code int, r render) {
	c.Lock()
	c.prevData.Code = code
	c.render = r
	c.stopHandle = true
	c.Unlock()
}

func (r *renderByte) Content(c *Context) []byte {
	if !c.hasContentType() {
		c.SetContentType(zutil.IfVal(r.Type != "", r.Type, ContentTypePlain).(string))
	}
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
	bf := zutil.GetBuff()
	defer zutil.PutBuff(bf)
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	if err := jsonEncoder.Encode(r.Data); err != nil {
		r.ContentDate, _ = json.Marshal(r.Data)
	} else {
		r.ContentDate = bf.Bytes()
	}

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
			isDebug := c.Engine.IsDebug()
			t = tpl.Get(isDebug)
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

func (c *Context) Byte(code int, value []byte) {
	c.renderProcessing(code, &renderByte{Data: value})
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.renderProcessing(code, &renderString{Format: format, Data: values})
}

// Deprecated: You can directly modify the return value of PrevContent()
func (c *Context) SetContent(data *PrevData) {
	c.Lock()
	c.prevData = data
	c.Unlock()
}

func (c *Context) File(path string) {
	fileExist := zfile.FileExist(path)
	code := zutil.IfVal(fileExist, 200, 404).(int)
	if fileExist {
		if f, err := os.Stat(path); err == nil {
			c.SetHeader("Last-Modified", f.ModTime().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT"))
		}
	}
	c.renderProcessing(code, &renderFile{Data: path, FileExist: fileExist})
}

func (c *Context) JSON(code int, values interface{}) {
	c.renderProcessing(code, &renderJSON{Data: values})
}

// ApiJSON ApiJSON
func (c *Context) ApiJSON(code int, msg string, data interface{}) {
	c.renderProcessing(http.StatusOK, &renderJSON{Data: ApiData{Code: code, Data: data,
		Msg: msg}})
}

// HTML export html
func (c *Context) HTML(code int, html string) {
	c.renderProcessing(code, &renderHTML{
		Data: html,
	})
}

// Template export tpl
func (c *Context) Template(code int, name string, data interface{}, funcMap ...map[string]interface{}) {
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
func (c *Context) Templates(code int, templates []string, data interface{}, funcMap ...map[string]interface{}) {
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

// Abort Abort
func (c *Context) Abort(code ...int) {
	c.Lock()
	c.stopHandle = true
	c.Unlock()
	if len(code) > 0 {
		c.SetStatus(code[0])
	}
	c.render = nil
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
	c.prevData.Code = code
	return c
}

func (c *Context) SetContentType(contentType string) *Context {
	c.SetHeader("Content-Type", contentType)
	return c
}

func (c *Context) hasContentType() bool {
	c.RLock()
	defer c.RUnlock()
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
	c.render = nil
	return c.prevData
}

func (t *tpl) Get(debug bool) *template.Template {
	if !debug {
		return t.tpl
	}
	tpl, _ := template.New("").Funcs(t.templateFuncMap).ParseGlob(t.pattern)
	return tpl
}
