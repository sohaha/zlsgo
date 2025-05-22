package znet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zutil"
)

type (
	// Renderer is the interface that wraps the Content method.
	// Any type implementing this interface can be used to render responses.
	Renderer interface {
		Content(c *Context) (content []byte)
	}

	// renderByte implements Renderer for raw byte data.
	renderByte struct {
		Data        []byte // Raw byte data to render
		Type        string // Content type
		ContentDate []byte // Cached content
	}

	// renderString implements Renderer for string data with formatting.
	renderString struct {
		Format      string        // Format string (printf style)
		Data        []interface{} // Format arguments
		ContentDate []byte        // Cached content
	}

	// renderJSON implements Renderer for JSON data.
	renderJSON struct {
		Data        interface{} // Data to be marshaled to JSON
		ContentDate []byte      // Cached content
	}

	// renderFile implements Renderer for file content.
	renderFile struct {
		Data        string // File path
		ContentDate []byte // Cached content
		FileExist   bool   // Whether the file exists
	}

	// renderHTML implements Renderer for HTML templates.
	renderHTML struct {
		Template    *template.Template // Parsed template
		Data        interface{}        // Template data
		ContentDate []byte             // Cached content
		FuncMap     template.FuncMap   // Template functions
		Templates   []string           // Template files
	}

	// ApiData represents a unified API response format with data, message, and status code.
	// It is used for standardizing JSON responses across the application.
	ApiData struct {
		Data interface{} `json:"data"`
		Msg  string      `json:"msg,omitempty"`
		Code int32       `json:"code" example:"200"`
	}

	render struct {
		data io.Writer
	}

	// Data is a convenience type for map[string]interface{} used for template data
	// and other data structures throughout the framework.
	Data map[string]interface{}

	// PrevData stores response information before it's sent to the client.
	// It includes status code, content type, and the actual content bytes.
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
	emptyBytes      = []byte{}
	bufferPool      = sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}
)

// renderProcessing handles the common rendering process for all renderer types.
// It sets the HTTP status code, processes the content, and writes it to the response.
func (c *Context) renderProcessing(code int32, r Renderer) {
	// if c.stopHandle.Load() && c.prevData.Code.Load() != 0 {
	// 	return
	// }
	if code != 0 {
		c.prevData.Code.Store(code)
	}
	c.mu.Lock()
	c.render = r
	c.mu.Unlock()
}

// Content implements the Renderer interface for renderByte.
// It returns the raw byte data and sets the appropriate content type.
func (r *renderByte) Content(c *Context) []byte {
	if !c.hasContentType() {
		c.SetContentType(ContentTypePlain)
	}
	return r.Data
}

// Content implements the Renderer interface for renderString.
// It formats the string data using the provided format and arguments.
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

// Content implements the Renderer interface for renderJSON.
// It marshals the data to JSON and sets the appropriate content type.
func (r *renderJSON) Content(c *Context) []byte {
	if r.ContentDate != nil {
		return r.ContentDate
	}
	c.SetContentType(ContentTypeJSON)
	r.ContentDate, _ = json.Marshal(r.Data)
	return r.ContentDate
}

// Content implements the Renderer interface for renderFile.
// It reads the file content and sets the appropriate content type based on file extension.
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

// Content implements the Renderer interface for renderHTML.
// It executes the template with the provided data and returns the rendered HTML.
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
		if c.Engine.views != nil {
			err = c.Engine.views.Render(&buf, r.Templates[0], r.Data)
		} else {
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
						Log.Error(err)
						return r.ContentDate
					}
				}
			}
			if t, err = templateParse(r.Templates, r.FuncMap); err == nil {
				err = t.Execute(&buf, r.Data)
			}
		}

		if err != nil {
			Log.Error(err)
		}
		r.ContentDate = buf.Bytes()
	} else {
		r.ContentDate = zstring.String2Bytes(fmt.Sprint(r.Data))
	}
	return r.ContentDate
}

// Byte writes raw bytes to the response with the given status code.
// It automatically detects the content type if possible.
func (c *Context) Byte(code int32, value []byte) {
	c.renderProcessing(code, &renderByte{Data: value})
}

// String writes a formatted string to the response with the given status code.
// It uses fmt.Sprintf-style formatting with the provided values.
func (c *Context) String(code int32, format string, values ...interface{}) {
	c.renderProcessing(code, &renderString{Format: format, Data: values})
}

func (r *render) Content(c *Context) (content []byte) {
	if r.data == nil {
		return emptyBytes
	}

	buf, ok := r.data.(*bytes.Buffer)
	if !ok {
		return emptyBytes
	}

	bufferPool.Put(buf)

	return buf.Bytes()
}

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
		if !isModified(c, f.ModTime()) {
			c.Abort(http.StatusNotModified)
			return
		}
	}
	c.renderProcessing(code, &renderFile{Data: path, FileExist: fileExist})
}

func (c *Context) JSON(code int32, values interface{}) {
	c.renderProcessing(code, &renderJSON{Data: values})
}

// ApiJSON ApiJSON
func (c *Context) ApiJSON(code int32, msg string, data interface{}) {
	c.renderProcessing(http.StatusOK, &renderJSON{Data: ApiData{
		Code: code, Data: data,
		Msg: msg,
	}})
}

// HTML export html
func (c *Context) HTML(code int32, html string) {
	c.renderProcessing(code, &renderHTML{
		Data: html,
	})
}

// GetWriter get render writer
func (c *Context) GetWriter(code int32) io.Writer {
	if !c.hasContentType() {
		c.SetContentType(ContentTypeHTML)
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()

	c.renderProcessing(code, &render{
		data: buf,
	})

	return buf
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

// IsAbort checks if the request handling has been aborted.
// It returns true if Abort() has been called, false otherwise.
func (c *Context) IsAbort() bool {
	return c.stopHandle.Load()
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

// SetStatus sets the HTTP status code for the response.
// It returns the context for method chaining.
func (c *Context) SetStatus(code int32) *Context {
	c.prevData.Code.Store(code)
	return c
}

// SetContentType sets the Content-Type header for the response.
// It returns the context for method chaining.
func (c *Context) SetContentType(contentType string) *Context {
	c.SetHeader("Content-Type", contentType)
	return c
}

// hasContentType checks if the Content-Type header has already been set.
// It returns true if the header exists, false otherwise.
func (c *Context) hasContentType() bool {
	r := c.mu.RLock()
	defer c.mu.RUnlock(r)
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
