package znet

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
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

func (c *Context) initPostForm() {
	if c.cacheForm != nil {
		return
	}
	form := make(url.Values)
	if c.Request.PostForm == nil {
		(func() {
			body, err := c.GetDataRaw()
			if err != nil {
				return
			}
			values, _ := url.ParseQuery(body)
			c.Request.PostForm = values
			v := c.ContentType()
			if v == mimeMultipartPOSTForm {
				_ = c.ParseMultipartForm()
			}
			form = c.Request.PostForm
		})()
	}
	c.cacheForm = form
}

// GetParam Get the value of the param inside the route
func (c *Context) GetParam(key string) string {
	return c.GetAllParam()[key]
}

// GetAllParam Get the value of all param in the route
func (c *Context) GetAllParam() map[string]string {
	if values, ok := c.Request.Context().Value(Utils.ContextKey).(map[string]string); ok {
		return values
	}

	return nil
}

// GetAllQuery Get All Queryst
func (c *Context) GetAllQuery() url.Values {
	c.initQuery()
	return c.cacheQuery
}

// GetAllQueryMaps Get All Queryst Maps
func (c *Context) GetAllQueryMaps() map[string]string {
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
	postForm := c.GetPostFormAll()
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
func (c *Context) GetPostFormAll() (value url.Values) {
	c.initPostForm()
	value = c.cacheForm
	return
}

// PostFormMap PostForm Map
func (c *Context) PostFormMap(key string) map[string]string {
	v, _ := c.GetPostFormMap(key)
	return v
}

// GetPostFormMap Get PostForm Map
func (c *Context) GetPostFormMap(key string) (map[string]string, bool) {
	postForm := c.GetPostFormAll()
	dicts, exist := c.get(postForm, key)
	if !exist && c.Request.MultipartForm != nil && c.Request.MultipartForm.File != nil {
		dicts, exist = c.get(c.Request.MultipartForm.Value, key)
	}

	return dicts, exist
}

// GetJSON Get JSON
func (c *Context) GetJSON(key string) *zjson.Res {
	j, _ := c.GetJSONs()

	return j.Get(key)
}

// GetJSONs Get JSONs
func (c *Context) GetJSONs() (json *zjson.Res, err error) {
	if c.cacheJSON != nil {
		return c.cacheJSON, nil
	}
	json = &zjson.Res{}
	var body string
	body, err = c.GetDataRaw()
	fmt.Println(len(body), err)
	if err != nil {
		return
	}
	if !zjson.Valid(body) {
		err = errors.New("illegal json format")
		return
	}

	json = zjson.Parse(body)
	c.cacheJSON = json
	return
}

// GetDataRaw Get Raw Data
func (c *Context) GetDataRaw() (string, error) {
	body, err := c.GetDataRawBytes()
	if err != nil {
		return "", err
	}
	return zstring.Bytes2String(body), err
}

// GetDataRawBytes Get Raw Data
func (c *Context) GetDataRawBytes() ([]byte, error) {
	if c.rawData != nil {
		return c.rawData, nil
	}
	var err error
	if c.Request.Body == nil {
		err = errors.New("request body is nil")
		return nil, err
	}
	var body []byte
	body, err = ioutil.ReadAll(c.Request.Body)
	if err == nil {
		c.rawData = body
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
	f, err := c.FormFiles(name)
	if err != nil {
		return nil, err
	}

	return f[0], err
}

// FormFiles Multiple FormFile
func (c *Context) FormFiles(name string) (files []*multipart.FileHeader, err error) {
	var multipartForm *multipart.Form
	multipartForm, err = c.MultipartForm()

	if multipartForm == nil || multipartForm.File == nil {
		err = errors.New("file is empty")
		return
	}

	files = make([]*multipart.FileHeader, 0, 1)
	if fhs := multipartForm.File[name]; len(fhs) > 0 {
		for i := range fhs {
			files = append(files, fhs[i])
		}
	}

	if len(files) == 0 {
		return nil, errors.New("file is empty")
	}

	return
}

// MultipartForm MultipartForm
func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.ParseMultipartForm()
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

func (c *Context) ParseMultipartForm(maxMultipartMemory ...int64) error {
	if c.Request.MultipartForm != nil {
		return nil
	}

	mr, err := c.multipartReader(false)
	if err != nil {
		return err
	}

	maxMemory := c.Engine.MaxMultipartMemory
	if len(maxMultipartMemory) > 0 && maxMultipartMemory[0] > 0 {
		maxMemory = maxMultipartMemory[0]
	}
	f, err := mr.ReadForm(maxMemory)
	if err != nil {
		return err
	}

	if c.Request.PostForm == nil {
		c.Request.PostForm = make(url.Values)
	}

	for k, v := range f.Value {
		c.Request.PostForm[k] = append(c.Request.PostForm[k], v...)
	}

	c.Request.MultipartForm = f
	return nil
}

func (c *Context) multipartReader(allowMixed bool) (*multipart.Reader, error) {
	v := c.Request.Header.Get("Content-Type")
	if v == "" {
		return nil, http.ErrNotMultipart
	}
	d, params, err := mime.ParseMediaType(v)
	if err != nil || !(d == "multipart/form-data" || allowMixed && d == "multipart/mixed") {
		return nil, http.ErrNotMultipart
	}
	boundary, ok := params["boundary"]
	if !ok {
		return nil, http.ErrMissingBoundary
	}
	body, err := c.GetDataRaw()
	if err != nil {
		return nil, err
	}
	return multipart.NewReader(strings.NewReader(body), boundary), nil
}
