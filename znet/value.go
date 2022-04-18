package znet

import (
	"errors"
	"reflect"

	"github.com/sohaha/zlsgo/zreflect"
	"github.com/sohaha/zlsgo/zvalid"
)

type setValueFunc func(kind reflect.Kind, field reflect.Value, fieldName, fieldTag string, value interface{}) error

type reqType uint8

// Content-Type MIME of the most common data formats
const (
	mimeJSON                      = "application/json"
	mimePlain                     = "text/plain"
	mimePOSTForm                  = "application/x-www-form-urlencoded"
	mimeMultipartPOSTForm         = "multipart/form-data"
	isJSON                reqType = iota
	isFrom
)

func (c *Context) valid(obj interface{}, v map[string]zvalid.Engine) error {
	r := make([]*zvalid.ValidEle, 0, len(v))
	val := reflect.ValueOf(obj)
	typ, err := zreflect.NewVal(val)
	val = val.Elem()
	if err != nil {
		return err
	}
	for k := range v {
		i, ok := typ.CheckExistsField(k)
		if !ok {
			continue
		}
		field := val.Field(i)
		switch field.Kind() {
		case reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
			value := field.Interface()
			r = append(r, zvalid.BatchVar(field, v[k].VerifiAny(value)))
		default:
			return errors.New("value validation for " + k + " is not supported")
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
