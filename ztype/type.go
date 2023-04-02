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

func (t Type) String() string {
	return ToString(t.v)
}

func (t Type) Bytes() []byte {
	return ToBytes(t.v)
}

func (t Type) Bool() bool {
	return ToBool(t.v)
}

func (t Type) Int() int {
	return ToInt(t.v)
}

func (t Type) Int8() int8 {
	return ToInt8(t.v)
}

func (t Type) Int16() int16 {
	return ToInt16(t.v)
}

func (t Type) Int32() int32 {
	return ToInt32(t.v)
}

func (t Type) Int64() int64 {
	return ToInt64(t.v)
}

func (t Type) Uint() uint {
	return ToUint(t.v)
}

func (t Type) Uint8() uint8 {
	return ToUint8(t.v)
}

func (t Type) Uint16() uint16 {
	return ToUint16(t.v)
}

func (t Type) Uint32() uint32 {
	return ToUint32(t.v)
}

func (t Type) Uint64() uint64 {
	return ToUint64(t.v)
}

func (t Type) Float32() float32 {
	return ToFloat32(t.v)
}

func (t Type) Float64() float64 {
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

func (t Type) Exists() bool {
	return t.v != nil
}

func (t Type) Time(format ...string) (time.Time, error) {
	return ToTime(t.v, format...)
}

func (t Type) Map() Map {
	return ToMap(t.v)
}
