package zvar

import (
	"os"
	"reflect"
)

// PathExists 路径是否存在
// 1存在并且是一个目录路径，2存在并且是一个文件路径，0不存在
func PathExists(path string) (int, error) {
	f, err := os.Stat(path)
	if err == nil {
		isFile := 2
		if f.IsDir() == true {
			isFile = 1
		}
		return isFile, nil
	}

	return 0, err
}

// GetType 获取变量类型
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
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() == reflect.Struct {
			varType = "struct"
		} else if v.Kind() == reflect.Invalid {
			varType = "interface{}"
		} else {
			varType = v.Type().String()
		}
	}
	return varType
}
