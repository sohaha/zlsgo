package ztype

import (
	"bytes"
	// "encoding/json"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/sohaha/zlsgo/zreflect"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztime"
)

// appString is an interface for types that can be converted to a string
// via their String() method.
type appString interface {
	String() string
}

// ToBytes converts any value to a byte slice.
// It first converts the value to a string using ToString and then converts the string to bytes.
func ToBytes(i interface{}) []byte {
	s := ToString(i)
	return zstring.String2Bytes(s)
}

// ToString converts any value to a string representation.
// It handles basic types directly and uses JSON marshaling for complex types.
// Returns an empty string if the input is nil.
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
		return strconv.FormatUint(value, 10)
	case float32:
		return strconv.FormatFloat(float64(value), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(value)
	case string:
		return value
	case []byte:
		return zstring.Bytes2String(value)
	default:
		if f, ok := value.(appString); ok {
			return f.String()
		}
		return toJsonString(value)
	}
}

// toJsonString converts a value to its JSON string representation.
// It removes surrounding quotes from the JSON output.
func toJsonString(value interface{}) string {
	jsonContent, _ := json.Marshal(value)
	jsonContent = bytes.Trim(jsonContent, `"`)
	return zstring.Bytes2String(jsonContent)
}

// ToBool converts any value to a boolean.
// Returns true for non-zero numbers, non-empty strings that aren't "false",
// and boolean true values. Returns false for everything else.
func ToBool(i interface{}) bool {
	if v, ok := i.(bool); ok {
		return v
	}
	if s := ToString(i); s != "" && s != "0" && !strings.EqualFold(s, "false") {
		return true
	}
	return false
}

// ToInt converts any value to an int.
// It uses ToInt64 internally and then converts the result to int.
func ToInt(i interface{}) int {
	if v, ok := i.(int); ok {
		return v
	}
	return int(ToInt64(i))
}

// ToInt8 converts any value to an int8.
// It uses ToInt64 internally and then converts the result to int8.
func ToInt8(i interface{}) int8 {
	if v, ok := i.(int8); ok {
		return v
	}
	return int8(ToInt64(i))
}

// ToInt16 converts any value to an int16.
// It uses ToInt64 internally and then converts the result to int16.
func ToInt16(i interface{}) int16 {
	if v, ok := i.(int16); ok {
		return v
	}
	return int16(ToInt64(i))
}

// ToInt32 converts any value to an int32.
// It uses ToInt64 internally and then converts the result to int32.
func ToInt32(i interface{}) int32 {
	if v, ok := i.(int32); ok {
		return v
	}
	return int32(ToInt64(i))
}

// ToInt64 converts any value to an int64.
// It handles numeric types directly and attempts to parse strings as integers.
// Supports decimal, hexadecimal (0x prefix), and octal (0 prefix) string formats.
// Returns 0 if the input is nil or cannot be converted.
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

// ToUint converts any value to a uint.
// It uses ToUint64 internally and then converts the result to uint.
func ToUint(i interface{}) uint {
	if v, ok := i.(uint); ok {
		return v
	}
	return uint(ToUint64(i))
}

// ToUint8 converts any value to a uint8.
// It uses ToUint64 internally and then converts the result to uint8.
func ToUint8(i interface{}) uint8 {
	if v, ok := i.(uint8); ok {
		return v
	}
	return uint8(ToUint64(i))
}

// ToUint16 converts any value to a uint16.
// It uses ToUint64 internally and then converts the result to uint16.
func ToUint16(i interface{}) uint16 {
	if v, ok := i.(uint16); ok {
		return v
	}
	return uint16(ToUint64(i))
}

// ToUint32 converts any value to a uint32.
// It uses ToUint64 internally and then converts the result to uint32.
func ToUint32(i interface{}) uint32 {
	if v, ok := i.(uint32); ok {
		return v
	}
	return uint32(ToUint64(i))
}

// ToUint64 converts any value to a uint64.
// It handles numeric types directly and attempts to parse strings as unsigned integers.
// Supports decimal, hexadecimal (0x prefix), and octal (0 prefix) string formats.
// Returns 0 if the input is nil or cannot be converted.
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

// ToFloat32 converts any value to a float32.
// It uses ToFloat64 internally and then converts the result to float32.
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

// ToFloat64 converts any value to a float64.
// It handles numeric types directly and attempts to parse strings as floating-point numbers.
// Returns 0.0 if the input is nil or cannot be converted.
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

// ToTime converts a value to a time.Time object.
// If the input is already a time.Time, it is returned directly.
// For string inputs, it attempts to parse using the provided format or a set of common formats.
// Returns the zero time and an error if the conversion fails.
func ToTime(i interface{}, format ...string) (time.Time, error) {
	switch val := i.(type) {
	case time.Time:
		return val, nil
	case int, int32, int64, uint, uint32, uint64:
		i := ToInt64(i)
		if i <= 9999999999 {
			return ztime.Unix(i), nil
		}
		if i <= 9999999999999 {
			i = i * 1000
		}
		return ztime.UnixMicro(i), nil
	default:
		if i := ToInt64(i); i > 0 {
			return ToTime(i)
		}
		v := ToString(i)
		return ztime.Parse(v, format...)
	}
}

// ToStruct converts a map or struct to another struct type.
// It uses reflection to match field names and performs appropriate type conversions.
// The outVal parameter must be a pointer to a struct.
// Returns an error if the conversion fails.
func ToStruct(v interface{}, outVal interface{}) error {
	val := zreflect.ValueOf(outVal)
	if reflect.Indirect(val).Kind() != reflect.Struct {
		return errors.New("result must be a struct")
	}
	return conv.to("", v, val, true)
}
