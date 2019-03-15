package gvar

import (
	"encoding/json"
	"strconv"
	"strings"
)

type appString interface {
	String() string
}

// ToByte 变量转[]byte
func ToByte(i interface{}) []byte {
	return []byte(ToString(i))
}

// ToString 变量转字符串
func ToString(i interface{}) string {
	if i == nil {
		return ""
	}
	switch value := i.(type) {
	case int:
		return strconv.Itoa(value)
	case int8:
		return strconv.Itoa(int(value))
	case int16:
		return strconv.Itoa(int(value))
	case int32:
		return strconv.Itoa(int(value))
	case int64:
		return strconv.Itoa(int(value))
	case uint:
		return strconv.FormatUint(uint64(value), 10)
	case uint8:
		return strconv.FormatUint(uint64(value), 10)
	case uint16:
		return strconv.FormatUint(uint64(value), 10)
	case uint32:
		return strconv.FormatUint(uint64(value), 10)
	case uint64:
		return strconv.FormatUint(uint64(value), 10)
	case float32:
		return strconv.FormatFloat(float64(value), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(value)
	case string:
		return value
	case []byte:
		return string(value)
	default:
		if f, ok := value.(appString); ok {
			return f.String()
		}
		jsonContent, _ := json.Marshal(value)
		return string(jsonContent)
	}
}

// ToBool 变量转布尔值
func ToBool(i interface{}) bool {
	if v, ok := i.(bool); ok {
		return v
	}
	if s := ToString(i); s != "" && s != "0" && s != "false" {
		return true
	}
	return false
}

// ToInt 变量转int
func ToInt(i interface{}) int {
	if v, ok := i.(int); ok {
		return v
	}
	return int(ToInt64(i))
}

// ToInt8 变量转int8
func ToInt8(i interface{}) int8 {
	if v, ok := i.(int8); ok {
		return v
	}
	return int8(ToInt64(i))
}

// ToInt16 变量转int16
func ToInt16(i interface{}) int16 {
	if v, ok := i.(int16); ok {
		return v
	}
	return int16(ToInt64(i))
}

// ToInt32 变量转int32
func ToInt32(i interface{}) int32 {
	if v, ok := i.(int32); ok {
		return v
	}
	return int32(ToInt64(i))
}

// ToInt64 变量转int64
func ToInt64(i interface{}) int64 {
	if i == nil {
		return 0
	}
	if v, ok := i.(int64); ok {
		return v
	}
	switch value := i.(type) {
	case int:
		return int64(value)
	case int8:
		return int64(value)
	case int16:
		return int64(value)
	case int32:
		return int64(value)
	// case int64:
	// 	return value
	case uint:
		return int64(value)
	case uint8:
		return int64(value)
	case uint16:
		return int64(value)
	case uint32:
		return int64(value)
	case uint64:
		return int64(value)
	case float32:
		return int64(value)
	case float64:
		return int64(value)
	case bool:
		if value {
			return 1
		}
		return 0
	default:
		s := ToString(value)
		if len(s) > 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
			if v, e := strconv.ParseInt(s[2:], 16, 64); e == nil {
				return v
			}
		}
		if len(s) > 1 && s[0] == '0' {
			if v, e := strconv.ParseInt(s[1:], 8, 64); e == nil {
				return v
			}
		}
		if v, e := strconv.ParseInt(s, 10, 64); e == nil {
			return v
		}
		return int64(ToFloat64(value))
	}
}

// ToUint 变量转uint
func ToUint(i interface{}) uint {
	if v, ok := i.(uint); ok {
		return v
	}
	return uint(ToUint64(i))
}

// ToUint8 变量转uint8
func ToUint8(i interface{}) uint8 {
	if v, ok := i.(uint8); ok {
		return v
	}
	return uint8(ToUint64(i))
}

// ToUint16 变量转uint16
func ToUint16(i interface{}) uint16 {
	if v, ok := i.(uint16); ok {
		return v
	}
	return uint16(ToUint64(i))
}

// ToUint32 变量转uint32
func ToUint32(i interface{}) uint32 {
	if v, ok := i.(uint32); ok {
		return v
	}
	return uint32(ToUint64(i))
}

// ToUint64 变量转uint64
func ToUint64(i interface{}) uint64 {
	if i == nil {
		return 0
	}
	switch value := i.(type) {
	case int:
		return uint64(value)
	case int8:
		return uint64(value)
	case int16:
		return uint64(value)
	case int32:
		return uint64(value)
	case int64:
		return uint64(value)
	case uint:
		return uint64(value)
	case uint8:
		return uint64(value)
	case uint16:
		return uint64(value)
	case uint32:
		return uint64(value)
	case uint64:
		return value
	case float32:
		return uint64(value)
	case float64:
		return uint64(value)
	case bool:
		if value {
			return 1
		}
		return 0
	default:
		s := ToString(value)
		if len(s) > 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
			if v, e := strconv.ParseUint(s[2:], 16, 64); e == nil {
				return v
			}
		}
		if len(s) > 1 && s[0] == '0' {
			if v, e := strconv.ParseUint(s[1:], 8, 64); e == nil {
				return v
			}
		}
		if v, e := strconv.ParseUint(s, 10, 64); e == nil {
			return v
		}
		return uint64(ToFloat64(value))
	}
}

// ToFloat32 变量转float32
func ToFloat32(i interface{}) float32 {
	if i == nil {
		return 0
	}
	if v, ok := i.(float32); ok {
		return v
	}
	v, _ := strconv.ParseFloat(strings.TrimSpace(ToString(i)), 64)
	return float32(v)
}

// ToFloat64 变量转float64
func ToFloat64(i interface{}) float64 {
	if i == nil {
		return 0
	}
	if v, ok := i.(float64); ok {
		return v
	}
	v, _ := strconv.ParseFloat(strings.TrimSpace(ToString(i)), 64)
	return v
}
