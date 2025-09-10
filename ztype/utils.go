package ztype

import (
	"reflect"
	"time"

	"github.com/sohaha/zlsgo/zreflect"
)

// GetType Get variable type
func GetType(s interface{}) string {
	var varType string
	switch s.(type) {
	case int:
		varType = "int"
	case int8:
		varType = "int8"
	case int16:
		varType = "int16"
	case int32:
		varType = "int32"
	case int64:
		varType = "int64"
	case uint:
		varType = "uint"
	case uint8:
		varType = "uint8"
	case uint16:
		varType = "uint16"
	case uint32:
		varType = "uint32"
	case uint64:
		varType = "uint64"
	case float32:
		varType = "float32"
	case float64:
		varType = "float64"
	case bool:
		varType = "bool"
	case string:
		varType = "string"
	case []byte:
		varType = "[]byte"
	default:
		if s == nil {
			return "nil"
		}
		v := zreflect.TypeOf(s)
		if v.Kind() == reflect.Invalid {
			return "invalid"
		}
		varType = v.String()
	}
	return varType
}

func reflectPtr(r reflect.Value) reflect.Value {
	if r.Kind() == reflect.Ptr {
		r = r.Elem()
	}
	return r
}

func parsePath(path string, v interface{}) (interface{}, bool) {
	return executeCompiledPath(compilePath(path), v)
}

var timeType = reflect.TypeOf(time.Time{})
