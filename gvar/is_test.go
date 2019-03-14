package gvar

import (
	"testing"
	. "github.com/sohaha/zlsgo/gtest"
)

func TestIs(t *testing.T) {

	var i int
	tIsInt := IsInt(i)
	equal(t, true, tIsInt)

	var i8 int8
	tIsInt8 := IsInt8(i8)
	equal(t, true, tIsInt8)

	var i16 int16
	tIsInt16 := IsInt16(i16)
	equal(t, true, tIsInt16)

	var i32 int32
	tIsInt32 := IsInt32(i32)
	equal(t, true, tIsInt32)

	var i64 int64
	tIsInt64 := IsInt64(i64)
	equal(t, true, tIsInt64)

	var ui uint
	tIsUint := IsUint(ui)
	equal(t, true, tIsUint)

	var ui8 uint8
	tIsUint8 := IsUint8(ui8)
	equal(t, true, tIsUint8)

	var ui16 uint16
	tIsUint16 := IsUint16(ui16)
	equal(t, true, tIsUint16)

	var ui32 uint32
	tIsUint32 := IsUint32(ui32)
	equal(t, true, tIsUint32)

	var ui64 uint64
	tIsUint64 := IsUint64(ui64)
	equal(t, true, tIsUint64)

	var f32 float32
	tIsFloat32 := IsFloat32(f32)
	equal(t, true, tIsFloat32)

	var f64 float64
	tIsFloat64 := IsFloat64(f64)
	equal(t, true, tIsFloat64)

	var bo bool
	tIsBool := IsBool(bo)
	equal(t, true, tIsBool)

	var str string
	tIsString := IsString(str)
	equal(t, true, tIsString)

	var by []byte
	tIsByte := IsByte(by)
	equal(t, true, tIsByte)

	type inTest interface {
	}

	type sutTest struct {
		test string
	}

	var in inTest
	tIsInterface := IsInterface(in)
	equal(t, true, tIsInterface)

	sut := sutTest{test: "T"}
	equal(t, true, IsStruct(sut))
	equal(t, true, IsStruct(&sut))

	m := map[string]string{}
	m["test"] = "testValue"
	tGetType1 := GetType(m)
	equal(t, "map[string]string", tGetType1)

	var m2 map[string]interface{}
	tGetType2 := GetType(m2)
	equal(t, "map[string]interface {}", tGetType2)

	var n chan int
	tGetType3 := GetType(n)
	equal(t, "chan int", tGetType3)

	dirPath := "."
	tIsDir := IsDir(dirPath)
	equal(t, true, tIsDir)

	filePath := "doc.go"
	tIsFile := IsFile(filePath)
	equal(t, true, tIsFile)

	notPath := "zls.php"
	status, _ := PathExists(notPath)
	equal(t, 0, status)
}
