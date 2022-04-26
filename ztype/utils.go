package ztype

import (
	"reflect"
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
		v := reflect.ValueOf(s)
		varType = v.Type().String()
	}
	return varType
}

func ReflectPtr(r reflect.Value) reflect.Value {
	if r.Kind() == reflect.Ptr {
		r = r.Elem()
	}
	return r
}

func InArray(needle, hystack interface{}) bool {
	nt := ToString(needle)
	for _, item := range Slice(hystack) {
		if nt == ToString(item) {
			return true
		}
	}
	return false
}
