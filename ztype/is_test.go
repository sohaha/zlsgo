package ztype

import (
	"testing"

	zls "github.com/sohaha/zlsgo"
)

func TestIs(t *testing.T) {

	T := zls.NewTest(t)
	var i int
	tIsInt := IsInt(i)
	T.Equal(true, tIsInt)

	var i8 int8
	tIsInt8 := IsInt8(i8)
	T.Equal(true, tIsInt8)

	var i16 int16
	tIsInt16 := IsInt16(i16)
	T.Equal(true, tIsInt16)

	var i32 int32
	tIsInt32 := IsInt32(i32)
	T.Equal(true, tIsInt32)

	var i64 int64
	tIsInt64 := IsInt64(i64)
	T.Equal(true, tIsInt64)

	var ui uint
	tIsUint := IsUint(ui)
	T.Equal(true, tIsUint)

	var ui8 uint8
	tIsUint8 := IsUint8(ui8)
	T.Equal(true, tIsUint8)

	var ui16 uint16
	tIsUint16 := IsUint16(ui16)
	T.Equal(true, tIsUint16)

	var ui32 uint32
	tIsUint32 := IsUint32(ui32)
	T.Equal(true, tIsUint32)

	var ui64 uint64
	tIsUint64 := IsUint64(ui64)
	T.Equal(true, tIsUint64)

	var f32 float32
	tIsFloat32 := IsFloat32(f32)
	T.Equal(true, tIsFloat32)

	var f64 float64
	tIsFloat64 := IsFloat64(f64)
	T.Equal(true, tIsFloat64)

	var bo bool
	tIsBool := IsBool(bo)
	T.Equal(true, tIsBool)

	var str string
	tIsString := IsString(str)
	T.Equal(true, tIsString)

	var by []byte
	tIsByte := IsByte(by)
	T.Equal(true, tIsByte)

	type inTest interface {
	}

	type sutTest struct {
		test string
	}

	var in inTest
	tIsInterface := IsInterface(in)
	T.Equal(true, tIsInterface)

	sut := sutTest{test: "T"}
	T.Equal(true, IsStruct(sut))
	T.Equal(true, IsStruct(&sut))
	T.Equal("ztype.sutTest", GetType(sut))
	T.Equal("*ztype.sutTest", GetType(&sut))

	m := map[string]string{}
	m["test"] = "testValue"
	tGetType1 := GetType(m)
	T.Equal("map[string]string", tGetType1)

	var m2 map[string]interface{}
	tGetType2 := GetType(m2)
	T.Equal("map[string]interface {}", tGetType2)

	var n chan int
	tGetType3 := GetType(n)
	T.Equal("chan int", tGetType3)
}
