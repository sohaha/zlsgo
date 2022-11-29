package ztype

import (
	"reflect"
	"strconv"
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
		if v.Kind() == reflect.Invalid {
			return "invalid"
		}
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

func parsePath(path string, v interface{}) (interface{}, bool) {
	t := 0
	i := 0
	val := v

	exist := true
	pp := func(p string, v interface{}) (result interface{}) {
		if v == nil || !exist {
			return nil
		}

		switch val := v.(type) {
		case Map:
			result, exist = val[p]
		case map[string]interface{}:
			result, exist = val[p]
		case map[string]string:
			result, exist = val[p]
		default:
			i, err := strconv.Atoi(p)
			if err == nil {
				switch val := v.(type) {
				case []Map:
					if len(val) > i {
						result = val[i]
					} else {
						exist = false
					}
				case []interface{}:
					if len(val) > i {
						result = val[i]
					} else {
						exist = false
					}
				case []string:
					if len(val) > i {
						result = val[i]
					} else {
						exist = false
					}
				}
			}
		}

		return
	}

	for ; i < len(path); i++ {
		switch path[i] {
		case '\\':
			ss := path[t:i]
			i++
			path = ss + path[i:]
		case '.':
			val = pp(path[t:i], val)
			t = i + 1
		}

		if exist == false {
			break
		}
	}

	if i != t {
		val = pp(path[t:], val)
	} else if i == 0 && t == 0 {
		val = pp(path, val)
	}

	return val, exist
}
