package znet

import (
	"errors"
	"reflect"

	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zutil"
	"github.com/sohaha/zlsgo/zvalid"
)

// Content-Type MIME of the most common data formats
const (
	MIMEJSON              = "application/json"
	MIMEPlain             = "text/plain"
	MIMEPOSTForm          = "application/x-www-form-urlencoded"
	MIMEMultipartPOSTForm = "multipart/form-data"
)

func (c *Context) Bind(obj interface{}) (err error) {
	return c.bind(obj, func(kind reflect.Kind, field reflect.Value, fieldTag string,
		value interface{}) error {
		return zutil.SetValue(kind, field, value)
	})
}

func (c *Context) BindValid(obj interface{}, elements map[string]zvalid.
Engine) (
	err error) {
	err = c.bind(obj, func(kind reflect.Kind, field reflect.Value,
		fieldTag string,
		value interface{}) error {
		validRule, ok := elements[fieldTag]
		if ok {
			delete(elements, fieldTag)
			var validValue string
			switch v := value.(type) {
			case string:
				validValue = v
			}
			value, err = validRule.Verifi(validValue).String()
			if err != nil {
				return err
			}
		}
		return zutil.SetValue(kind, field, value)
	})
	if err == nil {
		for _, v := range elements {
			err = v.Verifi("").Error()
			if err != nil {
				return err
			}
		}
	}
	return err
}

func (c *Context) bind(obj interface{}, set func(kind reflect.Kind,
	field reflect.Value, fieldTag string, value interface{}) error) (err error) {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr {
		err = errors.New("assignment requires the use of pointers")
		return
	}
	vv := v.Elem()
	tag := c.Engine.BindTag
	err = zutil.ReflectForNumField(vv, func(fieldTag string, kind reflect.Kind,
		field reflect.Value) error {
		var (
			value interface{}
			ok    bool
			err   error
		)
		// If you close the tag, the parameters will be transferred to SnakeCase by default
		if tag == "" {
			fieldTag = zstring.CamelCaseToSnakeCase(fieldTag)
		}
		if kind == reflect.Slice {
			value, ok = c.GetPostFormArray(fieldTag)
			if !ok {
				value, ok = c.GetQueryArray(fieldTag)
			}
		} else if kind == reflect.Struct {
			value, ok = c.GetPostFormMap(fieldTag)
		} else {
			value, ok = c.GetPostForm(fieldTag)
			if !ok {
				value, ok = c.GetQuery(fieldTag)
			}
		}
		if ok {
			err = set(kind, field, fieldTag, value)
		}
		// if err != nil {
		// 	err = fmt.Errorf("key: %s, %v", fieldTag, err)
		// }
		return err
	}, tag)
	return
}
