/*
 * @Author: seekwe
 * @Date:   2019-05-23 19:16:32
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-05-29 15:40:30
 */

package znet

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
)

type (
	render interface {
		Render(*Context, int) error
		// WriteContentType(w http.ResponseWriter)
	}
	renderString struct {
		Format string
		Data   []interface{}
	}
	renderJSON struct {
		Data interface{}
	}
	renderHTML struct {
		Template *template.Template
		Name     string
		Data     interface{}
	}
)

type H map[string]interface{}
type J struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func (c *Context) File(filepath string) {
	http.ServeFile(c.Writer, c.Request, filepath)
}

func (c *Context) render(code int, r render) {
	if !c.Info.StopHandle {
		if err := r.Render(c, code); err != nil {
			panic(err)
		}
	} else if c.Engine.webMode > releaseCode {
		c.Log.Warn("abort, not Render many times")
	}
}

var (
	plainContentType = "text/plain; charset=utf-8"
	htmlContentType  = "text/html; charset=utf-8"
	jsonContentType  = "application/json; charset=utf-8"
)

func writeContentType(w http.ResponseWriter, value string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		w.Header().Set("Content-Type", value)
	}
}

func (r renderString) Render(c *Context, code int) error {
	w := c.Writer
	writeContentType(w, plainContentType)
	c.StatusCode(code)
	if len(r.Data) > 0 {
		fmt.Fprintf(w, r.Format, r.Data...)
	} else {
		io.WriteString(w, r.Format)
	}
	return nil
}

func (r renderJSON) Render(c *Context, code int) error {
	w := c.Writer
	writeContentType(w, jsonContentType)
	c.StatusCode(code)
	jsonBytes, err := json.Marshal(r.Data)
	if err != nil {
		return err
	}
	w.Write(jsonBytes)
	return nil
}

func (r renderHTML) Render(c *Context, code int) (err error) {
	w := c.Writer
	writeContentType(w, htmlContentType)
	c.StatusCode(code)
	if r.Name != "" {
		t := template.Must(template.ParseFiles(r.Name))
		err = t.Execute(w, r.Data)
	} else {
		_, err = fmt.Fprint(c.Writer, r.Data)
	}
	return
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.render(code, renderString{Format: format, Data: values})
}

func (c *Context) JSON(code int, values interface{}) {
	c.render(code, renderJSON{Data: values})
}

func (c *Context) HTML(code int, html string) {
	c.render(code, renderHTML{
		Name: "",
		Data: html,
	})
}

func (c *Context) Template(code int, name string, data ...interface{}) {
	var _data interface{}
	if len(data) > 0 {
		_data = data[0]
	}
	c.render(code, renderHTML{
		Name: name,
		Data: _data,
	})
}

// Abort Abort
func (c *Context) Abort(code ...int) {
	c.Info.StopHandle = true
	if len(code) > 0 {
		c.StatusCode(code[0])
	}
}

// Redirect Redirect
func (c *Context) Redirect(link string, statusCode ...int) {
	c.Writer.Header().Set("Location", c.CompletionLink(link))
	code := 302
	if len(statusCode) > 0 {
		code = statusCode[0]
	}
	c.StatusCode(code)
}

// StatusCode StatusCode
func (c *Context) StatusCode(statusCode int) {
	c.Code = statusCode
	c.Writer.WriteHeader(c.Code)
}
