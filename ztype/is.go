package ztype

import (
	"reflect"

	"github.com/sohaha/zlsgo/zreflect"
)

// IsByte Is []byte
func IsByte(v interface{}) bool {
	return GetType(v) == "[]byte"
}

// IsString Is String
func IsString(v interface{}) bool {
	return GetType(v) == "string"
}

// IsBool Is Bool
func IsBool(v interface{}) bool {
	return GetType(v) == "bool"
}

// IsFloat64 Is float64
func IsFloat64(v interface{}) bool {
	return GetType(v) == "float64"
}

// IsFloat32 Is float32
func IsFloat32(v interface{}) bool {
	return GetType(v) == "float32"
}

// IsUint64 Is uint64
func IsUint64(v interface{}) bool {
	return GetType(v) == "uint64"
}

// IsUint32 Is uint32
func IsUint32(v interface{}) bool {
	return GetType(v) == "uint32"
}

// IsUint16 Is uint16
func IsUint16(v interface{}) bool {
	return GetType(v) == "uint16"
}

// IsUint8 Is uint8
func IsUint8(v interface{}) bool {
	return GetType(v) == "uint8"
}

// IsUint Is uint
func IsUint(v interface{}) bool {
	return GetType(v) == "uint"
}

// IsInt64 Is int64
func IsInt64(v interface{}) bool {
	return GetType(v) == "int64"
}

// IsInt32 Is int32
func IsInt32(v interface{}) bool {
	return GetType(v) == "int32"
}

// IsInt16 Is int16
func IsInt16(v interface{}) bool {
	return GetType(v) == "int16"
}

// IsInt8 Is int8
func IsInt8(v interface{}) bool {
	return GetType(v) == "int8"
}

// IsInt Is int
func IsInt(v interface{}) bool {
	return GetType(v) == "int"
}

// IsStruct Is Struct
func IsStruct(v interface{}) bool {
	r := reflectPtr(zreflect.ValueOf(v))
	return r.Kind() == reflect.Struct
}

// IsInterface Is interface{}
func IsInterface(v interface{}) bool {
	r := reflectPtr(zreflect.ValueOf(v))
	return r.Kind() == reflect.Invalid
}

func IsEmpty(value interface{}) bool {
	if value == nil {
		return true
	}
	switch value := value.(type) {
	case int:
		return value == 0
	case int8:
		return value == 0
	case int16:
		return value == 0
	case int32:
		return value == 0
	case int64:
		return value == 0
	case uint:
		return value == 0
	case uint8:
		return value == 0
	case uint16:
		return value == 0
	case uint32:
		return value == 0
	case uint64:
		return value == 0
	case float32:
		return value == 0
	case float64:
		return value == 0
	case bool:
		return !value
	case string:
		return value == ""
	case []byte:
		return len(value) == 0
	default:
		// Finally using reflect.
		rv := zreflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Chan,
			reflect.Map,
			reflect.Slice,
			reflect.Array:
			return rv.Len() == 0

		case reflect.Func,
			reflect.Ptr,
			reflect.Interface,
			reflect.UnsafePointer:
			if rv.IsNil() {
				return true
			}
		}
	}
	return false
}
