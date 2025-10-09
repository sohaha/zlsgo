package ztype

import (
	"reflect"

	"github.com/sohaha/zlsgo/zreflect"
)

// IsByte Is []byte
func IsByte(v interface{}) bool {
	_, ok := v.([]byte)
	return ok
}

// IsString Is String
func IsString(v interface{}) bool {
	_, ok := v.(string)
	return ok
}

// IsBool Is Bool
func IsBool(v interface{}) bool {
	_, ok := v.(bool)
	return ok
}

// IsFloat64 Is float64
func IsFloat64(v interface{}) bool {
	_, ok := v.(float64)
	return ok
}

// IsFloat32 Is float32
func IsFloat32(v interface{}) bool {
	_, ok := v.(float32)
	return ok
}

// IsUint64 Is uint64
func IsUint64(v interface{}) bool {
	_, ok := v.(uint64)
	return ok
}

// IsUint32 Is uint32
func IsUint32(v interface{}) bool {
	_, ok := v.(uint32)
	return ok
}

// IsUint16 Is uint16
func IsUint16(v interface{}) bool {
	_, ok := v.(uint16)
	return ok
}

// IsUint8 Is uint8
func IsUint8(v interface{}) bool {
	_, ok := v.(uint8)
	return ok
}

// IsUint Is uint
func IsUint(v interface{}) bool {
	_, ok := v.(uint)
	return ok
}

// IsInt64 Is int64
func IsInt64(v interface{}) bool {
	_, ok := v.(int64)
	return ok
}

// IsInt32 Is int32
func IsInt32(v interface{}) bool {
	_, ok := v.(int32)
	return ok
}

// IsInt16 Is int16
func IsInt16(v interface{}) bool {
	_, ok := v.(int16)
	return ok
}

// IsInt8 Is int8
func IsInt8(v interface{}) bool {
	_, ok := v.(int8)
	return ok
}

// IsInt Is int
func IsInt(v interface{}) bool {
	_, ok := v.(int)
	return ok
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

// IsEmpty checks if the given value is considered empty.
// Returns true for nil, zero values of basic types, empty strings, empty slices,
// empty arrays, empty maps, empty channels, nil pointers, and nil interfaces.
// Optimized to reduce reflection usage by checking most common types first.
func IsEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	switch value := value.(type) {
	case string:
		return value == ""
	case int:
		return value == 0
	case bool:
		return !value
	case []byte:
		return len(value) == 0
	case int64:
		return value == 0
	case float64:
		return value == 0
	case []interface{}:
		return len(value) == 0
	case map[string]interface{}:
		return len(value) == 0
	case []string:
		return len(value) == 0
	case []int:
		return len(value) == 0
	case int32:
		return value == 0
	case uint64:
		return value == 0
	case float32:
		return value == 0
	case uint:
		return value == 0
	case int16:
		return value == 0
	case int8:
		return value == 0
	case uint32:
		return value == 0
	case uint16:
		return value == 0
	case uint8:
		return value == 0
	default:
		rv := zreflect.ValueOf(value)
		kind := rv.Kind()

		switch kind {
		case reflect.Slice, reflect.Array, reflect.Map, reflect.Chan:
			return rv.Len() == 0
		case reflect.Ptr, reflect.Interface, reflect.UnsafePointer:
			return rv.IsNil()
		case reflect.Func:
			return rv.IsNil()
		case reflect.String:
			return rv.String() == ""
		case reflect.Bool:
			return !rv.Bool()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return rv.Int() == 0
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return rv.Uint() == 0
		case reflect.Float32, reflect.Float64:
			return rv.Float() == 0
		}
	}
	return false
}
