package znet

import (
	"errors"
	"reflect"

	"github.com/sohaha/zlsgo/zutil"
)

// Content-Type MIME of the most common data formats
const (
	MIMEJSON              = "application/json"
	MIMEPlain             = "text/plain"
	MIMEPOSTForm          = "application/x-www-form-urlencoded"
	MIMEMultipartPOSTForm = "multipart/form-data"
)

func (c *Context) Bind(obj interface{}) (err error) {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr {
		err = errors.New("assignment requires the use of pointers")
		return
	}
	vv := v.Elem()
	zutil.ReflectForNumField(vv, func(fieldTag string, kind reflect.Kind, field reflect.Value) bool {
		var (
			value interface{}
			ok    bool
		)
		if kind == reflect.Slice {
			value, ok = c.GetPostFormArray(fieldTag)
		} else if kind == reflect.Struct {
			value, ok = c.GetPostFormMap(fieldTag)
		} else {
			value, ok = c.GetPostForm(fieldTag)
		}
		if ok {
			err = zutil.SetValue(kind, field, value)
		}
		return err == nil
	})
	return
}
