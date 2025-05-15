// Package ztype provides flexible type conversion utilities and a dynamic type system
// that allows safe access to values with automatic type conversion.
package ztype

import (
	"time"
)

// Type is a wrapper around any value that provides safe type conversion methods.
// It allows accessing and converting values without having to handle type assertions
// and conversion errors manually.
type Type struct {
	v interface{}
}

// New creates a new Type instance wrapping the provided value.
// If the provided value is already a Type, it is returned as is.
func New(v interface{}) Type {
	switch val := v.(type) {
	case Type:
		return val
	default:
		return Type{v: v}
	}
}

// Value returns the underlying value stored in the Type wrapper.
func (t Type) Value() interface{} {
	return t.v
}

// Get retrieves a nested value using a path expression.
// Path expressions can navigate through maps and slices using dot notation and array indices.
// For example: "user.addresses[0].street"
// Returns an empty Type if the path doesn't exist.
func (t Type) Get(path string) Type {
	v, ok := parsePath(path, t.v)
	if !ok {
		return Type{}
	}
	return New(v)
}

// String converts the underlying value to a string.
// If the value is nil and a default value is provided, the default is returned.
func (t Type) String(def ...string) string {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToString(t.v)
}

// Bytes converts the underlying value to a byte slice.
// If the value is nil and a default value is provided, the default is returned.
func (t Type) Bytes(def ...[]byte) []byte {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToBytes(t.v)
}

// Bool converts the underlying value to a boolean.
// If the value is nil and a default value is provided, the default is returned.
func (t Type) Bool(def ...bool) bool {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToBool(t.v)
}

// Int converts the underlying value to an int.
// If the value is nil and a default value is provided, the default is returned.
func (t Type) Int(def ...int) int {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToInt(t.v)
}

// Int8 converts the underlying value to an int8.
// If the value is nil and a default value is provided, the default is returned.
func (t Type) Int8(def ...int8) int8 {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToInt8(t.v)
}

// Int16 converts the underlying value to an int16.
// If the value is nil and a default value is provided, the default is returned.
func (t Type) Int16(def ...int16) int16 {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToInt16(t.v)
}

// Int32 converts the underlying value to an int32.
// If the value is nil and a default value is provided, the default is returned.
func (t Type) Int32(def ...int32) int32 {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToInt32(t.v)
}

// Int64 converts the underlying value to an int64.
// If the value is nil and a default value is provided, the default is returned.
func (t Type) Int64(def ...int64) int64 {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToInt64(t.v)
}

// Uint converts the underlying value to a uint.
// If the value is nil and a default value is provided, the default is returned.
func (t Type) Uint(def ...uint) uint {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToUint(t.v)
}

// Uint8 converts the underlying value to a uint8.
// If the value is nil and a default value is provided, the default is returned.
func (t Type) Uint8(def ...uint8) uint8 {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToUint8(t.v)
}

// Uint16 converts the underlying value to a uint16.
// If the value is nil and a default value is provided, the default is returned.
func (t Type) Uint16(def ...uint16) uint16 {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToUint16(t.v)
}

// Uint32 converts the underlying value to a uint32.
// If the value is nil and a default value is provided, the default is returned.
func (t Type) Uint32(def ...uint32) uint32 {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToUint32(t.v)
}

// Uint64 converts the underlying value to a uint64.
// If the value is nil and a default value is provided, the default is returned.
func (t Type) Uint64(def ...uint64) uint64 {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToUint64(t.v)
}

// Float32 converts the underlying value to a float32.
// If the value is nil and a default value is provided, the default is returned.
func (t Type) Float32(def ...float32) float32 {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToFloat32(t.v)
}

// Float64 converts the underlying value to a float64.
// If the value is nil and a default value is provided, the default is returned.
func (t Type) Float64(def ...float64) float64 {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToFloat64(t.v)
}

// Maps converts the underlying value to a slice of maps.
// This is useful for handling JSON arrays of objects.
func (t Type) Maps() Maps {
	return t.Slice().Maps()
}

// Slice converts the underlying value to a SliceType.
// If noConv is true, it will not attempt to convert non-slice values to slices.
// Returns an empty SliceType if the value is nil or cannot be converted.
func (t Type) Slice(noConv ...bool) SliceType {
	if t.v == nil {
		return SliceType{}
	}
	return ToSlice(t.v, noConv...)
}

// SliceValue converts the underlying value to a slice of interface{} values.
// If noConv is true, it will not attempt to convert non-slice values to slices.
func (t Type) SliceValue(noConv ...bool) []interface{} {
	return t.Slice(noConv...).Value()
}

// SliceString converts the underlying value to a slice of strings.
// If noConv is true, it will not attempt to convert non-slice values to slices.
func (t Type) SliceString(noConv ...bool) []string {
	return t.Slice(noConv...).String()
}

// SliceInt converts the underlying value to a slice of integers.
// If noConv is true, it will not attempt to convert non-slice values to slices.
func (t Type) SliceInt(noConv ...bool) []int {
	return t.Slice(noConv...).Int()
}

// Exists checks if the underlying value is non-nil.
// Returns true if the value exists, false otherwise.
func (t Type) Exists() bool {
	return t.v != nil
}

// Time converts the underlying value to a time.Time.
// If the value is a string, the optional format parameter specifies the expected time format.
// Returns the converted time and any error that occurred during conversion.
func (t Type) Time(format ...string) (time.Time, error) {
	return ToTime(t.v, format...)
}

// Map converts the underlying value to a Map type.
// This is useful for handling JSON objects with dynamic access to properties.
func (t Type) Map() Map {
	return ToMap(t.v)
}
