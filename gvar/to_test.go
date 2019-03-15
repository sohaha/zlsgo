package gvar

import (
	"testing"

	. "github.com/sohaha/zlsgo/gtest"
)

type st interface {
	String() string
	Set(string)
}

var ni interface{}

type j struct {
	Name string
	Key  string
	Age  int `json:"age"`
}

var (
	str          = "123"
	i            = 123
	i8   int8    = 123
	i16  int16   = 123
	i32  int32   = 123
	i64  int64   = 123
	ui8  uint8   = 123
	ui   uint    = 123
	ui16 uint16  = 123
	ui32 uint32  = 123
	ui64 uint64  = 123
	f3   float32 = 123
	f6   float64 = 123
	b            = true
)

func (s *j) String() string {
	return ToString(s.Key)
}

func (s *j) Set(v string) {
	s.Key = v
}

func TestTo(t *testing.T) {
	var sst st = new(j)
	sst.Set(str)
	jj := j{Name: "123"}

	Equal(t, []byte(str), ToByte(str))
	Equal(t, []byte(str), ToByte(i))

	Equal(t, 0, ToInt(ni))
	Equal(t, i, ToInt(str))
	Equal(t, i, ToInt(i))
	Equal(t, i8, ToInt8(str))
	Equal(t, i8, ToInt8(i8))
	Equal(t, i16, ToInt16(str))
	Equal(t, i16, ToInt16(i16))
	Equal(t, i32, ToInt32(str))
	Equal(t, i32, ToInt32(i32))

	Equal(t, i64, ToInt64(str))
	Equal(t, i64, ToInt64(i))
	Equal(t, i64, ToInt64(i8))
	Equal(t, i64, ToInt64(i16))
	Equal(t, i64, ToInt64(i32))
	Equal(t, i64, ToInt64(i64))
	Equal(t, i64, ToInt64(ui8))
	Equal(t, i64, ToInt64(ui))
	Equal(t, i64, ToInt64(ui16))
	Equal(t, i64, ToInt64(ui32))
	Equal(t, i64, ToInt64(ui64))
	Equal(t, i64, ToInt64(f3))
	Equal(t, i64, ToInt64(f6))
	// 无法转换直接换成0
	Equal(t, ToInt64(0), ToInt64(jj))
	Equal(t, i64, ToInt64("0x7b"))
	Equal(t, i64, ToInt64("0173"))
	Equal(t, ToInt64(1), ToInt64(b))
	Equal(t, ToInt64(0), ToInt64(false))

	Equal(t, ToUint(0), ToUint(ni))
	Equal(t, ui, ToUint(str))
	Equal(t, ui, ToUint(ui))
	Equal(t, ui8, ToUint8(str))
	Equal(t, ui8, ToUint8(ui8))
	Equal(t, ui16, ToUint16(str))
	Equal(t, ui16, ToUint16(ui16))
	Equal(t, ui32, ToUint32(str))
	Equal(t, ui32, ToUint32(ui32))

	Equal(t, ui64, ToUint64(i64))
	Equal(t, ui64, ToUint64(str))
	Equal(t, ui64, ToUint64(i))
	Equal(t, ui64, ToUint64(i8))
	Equal(t, ui64, ToUint64(i16))
	Equal(t, ui64, ToUint64(i32))
	Equal(t, ui64, ToUint64(ui))
	Equal(t, ui64, ToUint64(ui8))
	Equal(t, ui64, ToUint64(ui16))
	Equal(t, ui64, ToUint64(ui32))
	Equal(t, ui64, ToUint64(ui64))
	Equal(t, ui64, ToUint64(f3))
	Equal(t, ui64, ToUint64(f6))
	// 无法转换直接换成0
	Equal(t, ToUint64(0), ToUint64(jj))
	Equal(t, ui64, ToUint64("0x7b"))
	Equal(t, ui64, ToUint64("0173"))
	Equal(t, ToUint64(1), ToUint64(b))
	Equal(t, ToUint64(0), ToUint64(false))

	Equal(t, str, ToString(sst))
	Equal(t, "", ToString(ni))
	Equal(t, "true", ToString(b))
	Equal(t, str, ToString(str))
	Equal(t, str, ToString(i8))
	Equal(t, str, ToString(ui))
	Equal(t, str, ToString(i))
	Equal(t, str, ToString(i8))
	Equal(t, str, ToString(i16))
	Equal(t, str, ToString(i32))
	Equal(t, str, ToString(i64))
	Equal(t, str, ToString(ui8))
	Equal(t, str, ToString(ui16))
	Equal(t, str, ToString(ui32))
	Equal(t, str, ToString(ui64))
	Equal(t, str, ToString(f6))
	Equal(t, str, ToString(f3))
	Equal(t, str, ToString(ToByte(i)))
	Equal(t, "{\"Name\":\"123\",\"Key\":\"\",\"age\":0}", ToString(jj))

	Equal(t, f6, ToFloat64(i))
	Equal(t, f6, ToFloat64(f3))
	Equal(t, f6, ToFloat64(f6))
	Equal(t, ToFloat64(0), ToFloat64(ni))

	Equal(t, f3, ToFloat32(i))
	Equal(t, f3, ToFloat32(f3))
	Equal(t, f3, ToFloat32(f6))
	Equal(t, ToFloat32(0), ToFloat32(ni))

	Equal(t, true, ToBool(b))
	Equal(t, true, ToBool(str))
	Equal(t, false, ToBool(ni))

}
