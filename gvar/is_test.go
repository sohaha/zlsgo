package gvar

import (
	"testing"

	. "github.com/sohaha/zlsgo/gtest"
)

func TestIs(t *testing.T) {

	var i int
	tIsInt := IsInt(i)
	Equal(t, true, tIsInt)

	var i8 int8
	tIsInt8 := IsInt8(i8)
	Equal(t, true, tIsInt8)

	var i16 int16
	tIsInt16 := IsInt16(i16)
	Equal(t, true, tIsInt16)

	var i32 int32
	tIsInt32 := IsInt32(i32)
	Equal(t, true, tIsInt32)

	var i64 int64
	tIsInt64 := IsInt64(i64)
	Equal(t, true, tIsInt64)

	var ui uint
	tIsUint := IsUint(ui)
	Equal(t, true, tIsUint)

	var ui8 uint8
	tIsUint8 := IsUint8(ui8)
	Equal(t, true, tIsUint8)

	var ui16 uint16
	tIsUint16 := IsUint16(ui16)
	Equal(t, true, tIsUint16)

	var ui32 uint32
	tIsUint32 := IsUint32(ui32)
	Equal(t, true, tIsUint32)

	var ui64 uint64
	tIsUint64 := IsUint64(ui64)
	Equal(t, true, tIsUint64)

	var f32 float32
	tIsFloat32 := IsFloat32(f32)
	Equal(t, true, tIsFloat32)

	var f64 float64
	tIsFloat64 := IsFloat64(f64)
	Equal(t, true, tIsFloat64)

	var bo bool
	tIsBool := IsBool(bo)
	Equal(t, true, tIsBool)

	var str string
	tIsString := IsString(str)
	Equal(t, true, tIsString)

	var by []byte
	tIsByte := IsByte(by)
	Equal(t, true, tIsByte)

	type inTest interface {
	}

	type sutTest struct {
		test string
	}

	var in inTest
	tIsInterface := IsInterface(in)
	Equal(t, true, tIsInterface)

	sut := sutTest{test: "T"}
	Equal(t, true, IsStruct(sut))
	Equal(t, true, IsStruct(&sut))

	m := map[string]string{}
	m["test"] = "testValue"
	tGetType1 := GetType(m)
	Equal(t, "map[string]string", tGetType1)

	var m2 map[string]interface{}
	tGetType2 := GetType(m2)
	Equal(t, "map[string]interface {}", tGetType2)

	var n chan int
	tGetType3 := GetType(n)
	Equal(t, "chan int", tGetType3)

	dirPath := "."
	tIsDir := IsDir(dirPath)
	Equal(t, true, tIsDir)

	filePath := "../doc.go"
	tIsFile := IsFile(filePath)
	Equal(t, true, tIsFile)

	notPath := "zls.php"
	status, _ := PathExists(notPath)
	Equal(t, 0, status)
}
