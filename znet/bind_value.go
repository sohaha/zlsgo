package znet

import (
	"errors"
	"reflect"

	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/zreflect"
	"github.com/sohaha/zlsgo/ztype"
)

// Bind binds the request data to the provided object based on the request method and content type.
// For GET requests, it binds query parameters. For requests with JSON content type, it binds JSON data.
// Otherwise, it binds form data. This provides a convenient way to handle different request formats.
func (c *Context) Bind(obj interface{}) (err error) {
	method := c.Request.Method
	if method == "GET" {
		return c.BindQuery(obj)
	}
	contentType := c.ContentType()
	if contentType == c.ContentType(ContentTypeJSON) {
		return c.BindJSON(obj)
	}
	return c.BindForm(obj)
}

// BindJSON binds JSON request body data to the provided object.
// It reads the raw request body and unmarshals it into the given object.
func (c *Context) BindJSON(obj interface{}) error {
	body, err := c.GetDataRawBytes()
	if err != nil {
		return err
	}

	if len(body) == 0 {
		return errors.New("request body is empty")
	}

	if !zjson.ValidBytes(body) {
		return errors.New("invalid JSON format")
	}

	return zjson.Unmarshal(body, obj)
}

// BindQuery binds URL query parameters to the provided object.
// It maps query parameters to struct fields based on field tags.
// Struct fields, slices, and basic types are all handled appropriately.
func (c *Context) BindQuery(obj interface{}) (err error) {
	q := c.GetAllQueryMaps()
	typ := zreflect.TypeOf(obj)
	m := make(map[string]interface{}, len(q))
	err = zreflect.ForEach(typ, func(parent []string, index int, tag string, field reflect.StructField) error {
		kind := field.Type.Kind()
		if kind == reflect.Struct {
			m[tag] = c.QueryMap(tag)
		} else if kind == reflect.Slice {
			v, _ := c.GetQueryArray(tag)
			m[tag] = v
		} else {
			v, ok := q[tag]
			if ok {
				m[tag] = v
			}
		}

		return zreflect.SkipChild
	})
	if err != nil {
		return err
	}
	return ztype.ToStruct(m, obj)
}

// BindForm binds form data from the request to the provided object.
// It handles both regular form data and multipart form data, mapping form fields
// to struct fields based on field tags.
func (c *Context) BindForm(obj interface{}) error {
	q := c.GetPostFormAll()
	typ := zreflect.TypeOf(obj)
	m := make(map[string]interface{}, len(q))
	err := zreflect.ForEach(typ, func(parent []string, index int, tag string, field reflect.StructField) error {
		kind := field.Type.Kind()
		if kind == reflect.Struct {
			m[tag] = c.PostFormMap(tag)
		} else if kind == reflect.Slice {
			sliceTyp := field.Type.Elem().Kind()
			if sliceTyp == reflect.Struct {
				// TODO follow up support
			} else {
				m[tag], _ = q[tag]
			}
		} else {
			v, ok := q[tag]
			if ok {
				m[tag] = v[0]
			}
		}

		return zreflect.SkipChild
	})
	if err != nil {
		return err
	}
	return ztype.ToStruct(m, obj)
}
