/*
 * @Author: seekwe
 * @Date:   2019-06-05 11:38:38
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-06-05 12:16:25
 */

package znet

import (
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"strings"
)

// GetParam Get the value of the param inside the route
func (c *Context) GetParam(key string) string {
	return c.GetAllParam()[key]
}

// GetAllParam Get the value of all param in the route
func (c *Context) GetAllParam() ParamsMapType {
	if values, ok := c.Request.Context().Value(contextKey).(ParamsMapType); ok {
		return values
	}

	return nil
}

func (c *Context) GetAllQueryst() url.Values {
	return c.Request.URL.Query()
}

func (c *Context) GetQueryArray(key string) ([]string, bool) {
	if values, ok := c.Request.URL.Query()[key]; ok && len(values) > 0 {
		return values, true
	}
	return []string{}, false
}

func (c *Context) GetQuery(key string) (string, bool) {
	if values, ok := c.GetQueryArray(key); ok {
		return values[0], ok
	}
	return "", false
}

func (c *Context) DefaultQuery(key string, defaultValue string) string {
	if value, ok := c.GetQuery(key); ok {
		return value
	}
	return defaultValue
}

func (c *Context) DefaultPostForm(key, defaultValue string) string {
	if value, ok := c.GetPostForm(key); ok {
		return value
	}
	return defaultValue
}

func (c *Context) GetPostForm(key string) (string, bool) {
	if values, ok := c.GetPostFormArray(key); ok {
		return values[0], ok
	}
	return "", false
}

func (c *Context) GetPostFormArray(key string) ([]string, bool) {
	req := c.Request
	req.ParseForm()
	req.ParseMultipartForm(c.Engine.MaxMultipartMemory)
	if values := req.PostForm[key]; len(values) > 0 {
		return values, true
	}
	if req.MultipartForm != nil && req.MultipartForm.File != nil {
		if values := req.MultipartForm.Value[key]; len(values) > 0 {
			return values, true
		}
	}
	return []string{}, false
}

func (c *Context) PostFormMap(key string) map[string]string {
	dicts, _ := c.GetPostFormMap(key)
	return dicts
}

func (c *Context) GetPostFormMap(key string) (map[string]string, bool) {
	req := c.Request
	req.ParseForm()
	req.ParseMultipartForm(c.Engine.MaxMultipartMemory)
	dicts, exist := c.get(req.PostForm, key)

	if !exist && req.MultipartForm != nil && req.MultipartForm.File != nil {
		dicts, exist = c.get(req.MultipartForm.Value, key)
	}

	return dicts, exist
}

func (c *Context) get(m map[string][]string, key string) (map[string]string, bool) {
	dicts := make(map[string]string)
	exist := false
	for k, v := range m {
		if i := strings.IndexByte(k, '['); i >= 1 && k[0:i] == key {
			if j := strings.IndexByte(k[i+1:], ']'); j >= 1 {
				exist = true
				dicts[k[i+1:][:j]] = v[0]
			}
		}
	}
	return dicts, exist
}

func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	_, fh, err := c.Request.FormFile(name)
	return fh, err
}

func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.Request.ParseMultipartForm(c.Engine.MaxMultipartMemory)
	return c.Request.MultipartForm, err
}

func (c *Context) SaveUploadedFile(file *multipart.FileHeader, dist string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dist)
	if err != nil {
		return err
	}
	defer out.Close()

	io.Copy(out, src)
	return nil
}
