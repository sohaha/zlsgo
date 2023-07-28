package ztype

import (
	"time"
)

type Type struct {
	v interface{}
}

func New(v interface{}) Type {
	switch val := v.(type) {
	case Type:
		return Type{v: val}
	default:
		return Type{v: v}
	}
}

func (t Type) Value() interface{} {
	return t.v
}

func (t Type) Get(path string) Type {
	v, ok := parsePath(path, t.v)
	if !ok {
		return Type{}
	}
	return New(v)
}

func (t Type) String(def ...string) string {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToString(t.v)
}

func (t Type) Bytes(def ...[]byte) []byte {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToBytes(t.v)
}

func (t Type) Bool(def ...bool) bool {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToBool(t.v)
}

func (t Type) Int(def ...int) int {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToInt(t.v)
}

func (t Type) Int8(def ...int8) int8 {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToInt8(t.v)
}

func (t Type) Int16(def ...int16) int16 {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToInt16(t.v)
}

func (t Type) Int32(def ...int32) int32 {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToInt32(t.v)
}

func (t Type) Int64(def ...int64) int64 {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToInt64(t.v)
}

func (t Type) Uint(def ...uint) uint {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToUint(t.v)
}

func (t Type) Uint8(def ...uint8) uint8 {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToUint8(t.v)
}

func (t Type) Uint16(def ...uint16) uint16 {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToUint16(t.v)
}

func (t Type) Uint32(def ...uint32) uint32 {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToUint32(t.v)
}

func (t Type) Uint64(def ...uint64) uint64 {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToUint64(t.v)
}

func (t Type) Float32(def ...float32) float32 {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToFloat32(t.v)
}

func (t Type) Float64(def ...float64) float64 {
	if t.v == nil && len(def) > 0 {
		return def[0]
	}
	return ToFloat64(t.v)
}

func (t Type) MapString() map[string]interface{} {
	return ToMapString(t.v)
}

func (t Type) Slice() SliceType {
	if t.v == nil {
		return make([]Type, 0)
	}
	return ToSlice(t.v)
}

func (t Type) SliceMaps() Maps {
	return t.Slice().Maps()
}

func (t Type) SliceValue() []interface{} {
	return t.Slice().Value()
}

func (t Type) SliceString() []string {
	return t.Slice().String()
}

func (t Type) SliceInt() []int {
	return t.Slice().Int()
}

func (t Type) Exists() bool {
	return t.v != nil
}

func (t Type) Time(format ...string) (time.Time, error) {
	return ToTime(t.v, format...)
}

func (t Type) Map() Map {
	return ToMap(t.v)
}
