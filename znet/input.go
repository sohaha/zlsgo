package znet

import (
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/url"
	"os"
	"strings"

	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/zstring"
)

func (c *Context) initQuery() {
	if c.cacheQuery != nil {
		return
	}
	c.cacheQuery = c.Request.URL.Query()
}

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

// GetAllQueryst Get All Queryst
func (c *Context) GetAllQueryst() url.Values {
	c.initQuery()
	return c.cacheQuery
}

// GetAllQuerystMaps Get All Queryst Maps
func (c *Context) GetAllQuerystMaps() map[string]string {
	c.initQuery()
	arr := map[string]string{}
	for key, v := range c.cacheQuery {
		arr[key] = v[0]
	}
	return arr
}

// GetQueryArray Get Query Array
func (c *Context) GetQueryArray(key string) ([]string, bool) {
	c.initQuery()
	if values, ok := c.cacheQuery[key]; ok && len(values) > 0 {
		return values, true
	}
	return []string{}, false
}

// GetQuery Get Query
func (c *Context) GetQuery(key string) (string, bool) {
	if values, ok := c.GetQueryArray(key); ok {
		return values[0], ok
	}
	return "", false
}

// DefaultQuery Get Query Or Default
func (c *Context) DefaultQuery(key string, def string) string {
	if value, ok := c.GetQuery(key); ok {
		return value
	}
	return def
}

// GetQueryMap Get Query Map
func (c *Context) GetQueryMap(key string) (map[string]string, bool) {
	return c.get(c.cacheQuery, key)
}

// QueryMap Get Query Map
func (c *Context) QueryMap(key string) map[string]string {
	v, _ := c.get(c.cacheQuery, key)
	return v
}

// DefaultPostForm Get Form Or Default
func (c *Context) DefaultPostForm(key, def string) string {
	if value, ok := c.GetPostForm(key); ok {
		return value
	}
	return def
}

// GetPostForm Get PostForm
func (c *Context) GetPostForm(key string) (string, bool) {
	if values, ok := c.GetPostFormArray(key); ok {
		return values[0], ok
	}
	return "", false
}

// DefaultFormOrQuery  Get Form Or Query
func (c *Context) DefaultFormOrQuery(key string, def string) string {
	if value, ok := c.GetPostForm(key); ok {
		return value
	}
	return c.DefaultQuery(key, def)
}

// GetPostFormArray Get Post FormArray
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

// GetPostFormAll Get PostForm All
func (c *Context) GetPostFormAll() (value url.Values, err error) {
	req := c.Request
	if req.PostForm == nil {
		if c.ContentType() == mimeMultipartPOSTForm {
			err = req.ParseMultipartForm(c.Engine.MaxMultipartMemory)
		} else {
			err = req.ParseForm()
		}
	}
	value = req.PostForm
	return
}

// PostFormMap PostForm Map
func (c *Context) PostFormMap(key string) map[string]string {
	v, _ := c.GetPostFormMap(key)
	return v
}

// GetPostFormMap Get PostForm Map
func (c *Context) GetPostFormMap(key string) (map[string]string, bool) {
	req := c.Request
	postForm, _ := c.GetPostFormAll()
	dicts, exist := c.get(postForm, key)
	if !exist && req.MultipartForm != nil && req.MultipartForm.File != nil {
		dicts, exist = c.get(req.MultipartForm.Value, key)
	}

	return dicts, exist
}

// GetJSON Get JSON
func (c *Context) GetJSON(key string) zjson.Res {
	j, _ := c.GetJSONs()

	return j.Get(key)
}

// GetJSONs Get JSONs
func (c *Context) GetJSONs() (json zjson.Res, err error) {
	if c.cacheJSON != nil {
		return *c.cacheJSON, nil
	}
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
	c.cacheJSON = &json
	return
}

// GetDataRaw Get Raw Data
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
	d := make(map[string]string)
	e := false
	for k, v := range m {
		if i := strings.IndexByte(k, '['); i >= 1 && k[0:i] == key {
			if j := strings.IndexByte(k[i+1:], ']'); j >= 1 {
				e = true
				d[k[i+1:][:j]] = v[0]
			}
		}
	}
	return d, e
}

// FormFile FormFile
func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	_, _ = c.MultipartForm()
	_, fh, err := c.Request.FormFile(name)
	return fh, err
}

// MultipartForm MultipartForm
func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.Request.ParseMultipartForm(c.Engine.MaxMultipartMemory)
	return c.Request.MultipartForm, err
}

// SaveUploadedFile Save Uploaded File
func (c *Context) SaveUploadedFile(file *multipart.FileHeader, dist string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dist = zfile.RealPath(dist)
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
