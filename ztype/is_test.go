package ztype

import (
	"testing"
	"unsafe"

	"github.com/sohaha/zlsgo"
)

func TestIs(t *testing.T) {
	T := zlsgo.NewTest(t)
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

	type inTest interface{}

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

func BenchmarkIsStr(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = isStr("sss").(string)
	}
}

func BenchmarkZIsStr(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = IsString(isStr("sss"))
	}
}

func BenchmarkGetType(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetType(isStr("sss"))
	}
}

func isStr(s interface{}) interface{} {
	return s
}

func TestIsEmpty(T *testing.T) {
	t := zlsgo.NewTest(T)
	t.EqualTrue(IsEmpty(""))
	t.EqualTrue(IsEmpty(0))
	t.EqualTrue(IsEmpty([]byte("")))
	t.EqualTrue(IsEmpty(int(0)))
	t.EqualTrue(IsEmpty(int8(0)))
	t.EqualTrue(IsEmpty(int16(0)))
	t.EqualTrue(IsEmpty(int32(0)))
	t.EqualTrue(IsEmpty(int64(0)))
	t.EqualTrue(IsEmpty(uint(0)))
	t.EqualTrue(IsEmpty(uint8(0)))
	t.EqualTrue(IsEmpty(uint16(0)))
	t.EqualTrue(IsEmpty(uint32(0)))
	t.EqualTrue(IsEmpty(uint64(0)))
	t.EqualTrue(IsEmpty(float32(0)))
	t.EqualTrue(IsEmpty(float64(0)))
	t.EqualTrue(IsEmpty(false))
	t.EqualTrue(IsEmpty([]string{}))
	var s interface{}
	t.EqualTrue(IsEmpty(s))
}

func TestIsEmptyBroad(t *testing.T) {
	if !IsEmpty(nil) {
		t.Fatal("nil should be empty")
	}
	if !IsEmpty(0) || IsEmpty(1) {
		t.Fatal("int empty check failed")
	}
	if !IsEmpty(0.0) || IsEmpty(0.1) {
		t.Fatal("float64 empty check failed")
	}
	if !IsEmpty("") || IsEmpty("a") {
		t.Fatal("string empty check failed")
	}
	if !IsEmpty(false) || IsEmpty(true) {
		t.Fatal("bool empty check failed")
	}
	if !IsEmpty([]byte{}) || IsEmpty([]byte{1}) {
		t.Fatal("[]byte empty check failed")
	}
	if !IsEmpty([]interface{}{}) || IsEmpty([]interface{}{"x"}) {
		t.Fatal("[]interface{} empty check failed")
	}
	if !IsEmpty(map[string]interface{}{}) || IsEmpty(map[string]interface{}{"a": 1}) {
		t.Fatal("map[string]interface{} empty check failed")
	}
	if !IsEmpty([]string{}) || IsEmpty([]string{"a"}) {
		t.Fatal("[]string empty check failed")
	}
	if !IsEmpty([]int{}) || IsEmpty([]int{1}) {
		t.Fatal("[]int empty check failed")
	}
	if !IsEmpty(int32(0)) || IsEmpty(int32(1)) {
		t.Fatal("int32 empty check failed")
	}
	if !IsEmpty(uint64(0)) || IsEmpty(uint64(1)) {
		t.Fatal("uint64 empty check failed")
	}
	if !IsEmpty(float32(0)) || IsEmpty(float32(1)) {
		t.Fatal("float32 empty check failed")
	}
	if !IsEmpty(uint(0)) || IsEmpty(uint(1)) {
		t.Fatal("uint empty check failed")
	}
	if !IsEmpty(int16(0)) || IsEmpty(int16(1)) {
		t.Fatal("int16 empty check failed")
	}
	if !IsEmpty(int8(0)) || IsEmpty(int8(1)) {
		t.Fatal("int8 empty check failed")
	}
	if !IsEmpty(uint32(0)) || IsEmpty(uint32(1)) {
		t.Fatal("uint32 empty check failed")
	}
	if !IsEmpty(uint16(0)) || IsEmpty(uint16(1)) {
		t.Fatal("uint16 empty check failed")
	}
	if !IsEmpty(uint8(0)) || IsEmpty(uint8(1)) {
		t.Fatal("uint8 empty check failed")
	}

	ch := make(chan int)
	if !IsEmpty(ch) {
		t.Fatal("chan should be empty when len==0")
	}
	var fn func()
	if !IsEmpty(fn) {
		t.Fatal("nil func should be empty")
	}
	fn = func() {}
	if IsEmpty(fn) {
		t.Fatal("non-nil func should not be empty")
	}
	var pi *int
	if !IsEmpty(pi) {
		t.Fatal("nil pointer should be empty")
	}
	x := 0
	pi = &x
	if IsEmpty(pi) {
		t.Fatal("non-nil pointer should not be empty")
	}
}

func TestIsEmptyReflectionEdgeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

	arr := [3]int{0, 0, 0}
	tt.EqualTrue(!IsEmpty(arr))

	emptyArr := [0]int{}
	tt.EqualTrue(IsEmpty(emptyArr))

	var up unsafe.Pointer
	tt.EqualTrue(IsEmpty(up))

	type customInt int
	var ci customInt = 0
	tt.EqualTrue(IsEmpty(ci))

	ci = 5
	tt.EqualTrue(!IsEmpty(ci))
}

func TestGetTypeEdgeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Equal("nil", GetType(nil))
	tt.Equal("int", GetType(int(0)))
	tt.Equal("int8", GetType(int8(0)))
	tt.Equal("int16", GetType(int16(0)))
	tt.Equal("int32", GetType(int32(0)))
	tt.Equal("int64", GetType(int64(0)))
	tt.Equal("uint", GetType(uint(0)))
	tt.Equal("uint8", GetType(uint8(0)))
	tt.Equal("uint16", GetType(uint16(0)))
	tt.Equal("uint32", GetType(uint32(0)))
	tt.Equal("uint64", GetType(uint64(0)))
	tt.Equal("float32", GetType(float32(0)))
	tt.Equal("float64", GetType(float64(0)))
	tt.Equal("bool", GetType(bool(false)))
	tt.Equal("string", GetType(string("")))
	tt.Equal("[]byte", GetType([]byte{}))

	type myStruct struct{ X int }
	tt.Equal("ztype.myStruct", GetType(myStruct{}))
	tt.Equal("*ztype.myStruct", GetType(&myStruct{}))
}

func TestIsEmptyWithMoreTypes(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.EqualTrue(IsEmpty(int64(0)))
	tt.EqualTrue(!IsEmpty(int64(1)))
	tt.EqualTrue(IsEmpty(float64(0)))
	tt.EqualTrue(!IsEmpty(float64(0.1)))
	tt.EqualTrue(IsEmpty([]interface{}{}))
	tt.EqualTrue(!IsEmpty([]interface{}{1}))
	tt.EqualTrue(IsEmpty(map[string]interface{}{}))
	tt.EqualTrue(!IsEmpty(map[string]interface{}{"a": 1}))
	tt.EqualTrue(IsEmpty([]string{}))
	tt.EqualTrue(!IsEmpty([]string{"a"}))
	tt.EqualTrue(IsEmpty([]int{}))
	tt.EqualTrue(!IsEmpty([]int{1}))
}
