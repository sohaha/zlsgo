/*
 * @Author: seekwe
 * @Date:   2019-05-09 13:08:23
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-05-25 12:50:47
 */

package ztype

import (
	"reflect"
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
	r := reflectPtr(v)
	return r.Kind() == reflect.Struct
}

// IsInterface Is interface{}
func IsInterface(v interface{}) bool {
	r := reflectPtr(v)
	return r.Kind() == reflect.Invalid
}
