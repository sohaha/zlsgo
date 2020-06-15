package znet

import (
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/url"
	"os"
	"strings"

	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/zstring"
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

func (c *Context) GetAllQuerystMaps() map[string]string {
	arr := map[string]string{}
	for key, v := range c.Request.URL.Query() {
		arr[key] = v[0]
	}
	return arr
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

func (c *Context) DefaultQuery(key string, def string) string {
	if value, ok := c.GetQuery(key); ok {
		return value
	}
	return def
}

func (c *Context) DefaultPostForm(key, def string) string {
	if value, ok := c.GetPostForm(key); ok {
		return value
	}
	return def
}

func (c *Context) GetPostForm(key string) (string, bool) {
	if values, ok := c.GetPostFormArray(key); ok {
		return values[0], ok
	}
	return "", false
}

func (c *Context) DefaultFormOrQuery(key string, def string) string {
	if value, ok := c.GetPostForm(key); ok {
		return value
	}
	return c.DefaultQuery(key, def)
}

func (c *Context) GetPostFormArray(key string) ([]string, bool) {
	req := c.Request
	postForm, _ := c.GetPostFormAll()
	if values := postForm[key]; len(values) > 0 {
		return values, true
	}
	if req.MultipartForm != nil && req.MultipartForm.File != nil {
		if values := req.MultipartForm.Value[key]; len(values) > 0 {
			return values, true
		}
	}
	return []string{}, false
}

func (c *Context) GetPostFormAll() (value url.Values, err error) {
	req := c.Request
	if req.PostForm == nil {
		if c.ContentType() == MIMEMultipartPOSTForm {
			if c.Request.Method == "DELETE"{

			}
			err = req.ParseMultipartForm(c.Engine.MaxMultipartMemory)
		} else {
			err = req.ParseForm()
		}
	}
	value = req.PostForm
	return
}

func (c *Context) PostFormMap(key string) map[string]string {
	dicts, _ := c.GetPostFormMap(key)
	return dicts
}

func (c *Context) GetPostFormMap(key string) (map[string]string, bool) {
	req := c.Request
	postForm, _ := c.GetPostFormAll()
	dicts, exist := c.get(postForm, key)
	if !exist && req.MultipartForm != nil && req.MultipartForm.File != nil {
		dicts, exist = c.get(req.MultipartForm.Value, key)
	}

	return dicts, exist
}

func (c *Context) GetJSON(key string) zjson.Res {
	j, _ := c.GetJSONs()

	return j.Get(key)
}

func (c *Context) GetJSONs() (json zjson.Res, err error) {
	var body string
	body, err = c.GetDataRaw()
	if err != nil {
		return
	}
	if !zjson.Valid(body) {
		err = errors.New("illegal json format")
		return
	}
	json = zjson.Parse(body)
	return
}

func (c *Context) GetDataRaw() (string, error) {
	if c.rawData != "" {
		return c.rawData, nil
	}
	var err error
	if c.Request.Body == nil {
		err = errors.New("request.Body is nil")
		return "", err
	}
	var body []byte
	body, err = ioutil.ReadAll(c.Request.Body)
	if err == nil {
		c.rawData = zstring.Bytes2String(body)
	}
	return c.rawData, err
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
	_, err = io.Copy(out, src)
	if err != nil {
		return err
	}

	return nil
}
