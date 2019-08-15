package ztype

import (
	"testing"

	zls "github.com/sohaha/zlsgo"
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
	T := zls.NewTest(t)
	var sst st = new(j)
	sst.Set(str)
	jj := j{Name: "123"}

	T.Equal([]byte(str), ToByte(str))
	T.Equal([]byte(str), ToByte(i))

	T.Equal(0, ToInt(ni))
	T.Equal(i, ToInt(str))
	T.Equal(i, ToInt(i))
	T.Equal(i8, ToInt8(str))
	T.Equal(i8, ToInt8(i8))
	T.Equal(i16, ToInt16(str))
	T.Equal(i16, ToInt16(i16))
	T.Equal(i32, ToInt32(str))
	T.Equal(i32, ToInt32(i32))

	T.Equal(i64, ToInt64(str))
	T.Equal(i64, ToInt64(i))
	T.Equal(i64, ToInt64(i8))
	T.Equal(i64, ToInt64(i16))
	T.Equal(i64, ToInt64(i32))
	T.Equal(i64, ToInt64(i64))
	T.Equal(i64, ToInt64(ui8))
	T.Equal(i64, ToInt64(ui))
	T.Equal(i64, ToInt64(ui16))
	T.Equal(i64, ToInt64(ui32))
	T.Equal(i64, ToInt64(ui64))
	T.Equal(i64, ToInt64(f3))
	T.Equal(i64, ToInt64(f6))
	// 无法转换直接换成0
	T.Equal(ToInt64(0), ToInt64(jj))
	T.Equal(i64, ToInt64("0x7b"))
	T.Equal(i64, ToInt64("0173"))
	T.Equal(ToInt64(1), ToInt64(b))
	T.Equal(ToInt64(0), ToInt64(false))

	T.Equal(ToUint(0), ToUint(ni))
	T.Equal(ui, ToUint(str))
	T.Equal(ui, ToUint(ui))
	T.Equal(ui8, ToUint8(str))
	T.Equal(ui8, ToUint8(ui8))
	T.Equal(ui16, ToUint16(str))
	T.Equal(ui16, ToUint16(ui16))
	T.Equal(ui32, ToUint32(str))
	T.Equal(ui32, ToUint32(ui32))

	T.Equal(ui64, ToUint64(i64))
	T.Equal(ui64, ToUint64(str))
	T.Equal(ui64, ToUint64(i))
	T.Equal(ui64, ToUint64(i8))
	T.Equal(ui64, ToUint64(i16))
	T.Equal(ui64, ToUint64(i32))
	T.Equal(ui64, ToUint64(ui))
	T.Equal(ui64, ToUint64(ui8))
	T.Equal(ui64, ToUint64(ui16))
	T.Equal(ui64, ToUint64(ui32))
	T.Equal(ui64, ToUint64(ui64))
	T.Equal(ui64, ToUint64(f3))
	T.Equal(ui64, ToUint64(f6))
	// 无法转换直接换成0
	T.Equal(ToUint64(0), ToUint64(jj))
	T.Equal(ui64, ToUint64("0x7b"))
	T.Equal(ui64, ToUint64("0173"))
	T.Equal(ToUint64(1), ToUint64(b))
	T.Equal(ToUint64(0), ToUint64(false))

	T.Equal(str, ToString(sst))
	T.Equal("", ToString(ni))
	T.Equal("true", ToString(b))
	T.Equal(str, ToString(str))
	T.Equal(str, ToString(i8))
	T.Equal(str, ToString(ui))
	T.Equal(str, ToString(i))
	T.Equal(str, ToString(i8))
	T.Equal(str, ToString(i16))
	T.Equal(str, ToString(i32))
	T.Equal(str, ToString(i64))
	T.Equal(str, ToString(ui8))
	T.Equal(str, ToString(ui16))
	T.Equal(str, ToString(ui32))
	T.Equal(str, ToString(ui64))
	T.Equal(str, ToString(f6))
	T.Equal(str, ToString(f3))
	T.Equal(str, ToString(ToByte(i)))
	T.Equal("{\"Name\":\"123\",\"Key\":\"\",\"age\":0}", ToString(jj))
	T.Equal(f6, ToFloat64(i))
	T.Equal(f6, ToFloat64(f3))
	T.Equal(f6, ToFloat64(f6))
	T.Equal(ToFloat64(0), ToFloat64(ni))

	T.Equal(f3, ToFloat32(i))
	T.Equal(f3, ToFloat32(f3))
	T.Equal(f3, ToFloat32(f6))
	T.Equal(ToFloat32(0), ToFloat32(ni))

	T.Equal(true, ToBool(b))
	T.Equal(true, ToBool(str))
	T.Equal(false, ToBool(ni))

}
