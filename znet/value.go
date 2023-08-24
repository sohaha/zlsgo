package znet

import (
	"errors"
	"reflect"

	"github.com/sohaha/zlsgo/zreflect"
	"github.com/sohaha/zlsgo/zvalid"
)

// Content-Type MIME of the most common data formats
const (
	mimeJSON              = "application/json"
	mimePlain             = "text/plain"
	mimePOSTForm          = "application/x-www-form-urlencoded"
	mimeMultipartPOSTForm = "multipart/form-data"
)

func (c *Context) valid(obj interface{}, v map[string]zvalid.Engine) error {
	r := make([]*zvalid.ValidEle, 0, len(v))
	val := zreflect.ValueOf(obj)
	if val.Kind() != reflect.Ptr {
		return errors.New("result must be a pointer")
	}

	val = val.Elem()
	typ := zreflect.TypeOf(val)
	for i := 0; i < typ.NumField(); i++ {
		field := val.Field(i)
		name, _ := zreflect.GetStructTag(typ.Field(i))
		switch field.Kind() {
		case reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
			value := field.Interface()
			if rv, ok := v[name]; ok {
				r = append(r, zvalid.BatchVar(field, rv.VerifiAny(value)))
			}
		case reflect.Struct:
		case reflect.Slice:
		default:
			return errors.New("value validation for " + name + " is not supported")
		}
	}

	return zvalid.Batch(r...)
}

func (c *Context) BindValid(obj interface{}, v map[string]zvalid.Engine) error {
	err := c.Bind(obj)
	if err != nil {
		return err
	}

	return c.valid(obj, v)
}
