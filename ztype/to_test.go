package ztype_test

import (
	"fmt"
	"strconv"
	"testing"

	zls "github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/ztype"
)

type st interface {
	String() string
	Set(string)
}

type (
	type1 struct {
		B  string
		A  int
		C1 float32
	}
	type2 struct {
		E  *uint
		G  map[string]int `z:"gg"`
		S2 *type1
		F  []string `json:"fs"`
		type1
		S1 type1
		D  bool
	}
)

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
	return ztype.ToString(s.Key)
}

func (s *j) Set(v string) {
	s.Key = v
}

func TestTo(t *testing.T) {
	tt := zls.NewTest(t)
	var sst st = new(j)
	sst.Set(str)
	jj := j{Name: "123"}

	tt.Equal([]byte(str), ztype.ToBytes(str))
	tt.Equal([]byte(str), ztype.ToBytes(i))

	tt.Equal(0, ztype.ToInt(ni))
	tt.Equal(i, ztype.ToInt(str))
	tt.Equal(i, ztype.ToInt(i))
	tt.Equal(i8, ztype.ToInt8(str))
	tt.Equal(i8, ztype.ToInt8(i8))
	tt.Equal(i16, ztype.ToInt16(str))
	tt.Equal(i16, ztype.ToInt16(i16))
	tt.Equal(i32, ztype.ToInt32(str))
	tt.Equal(i32, ztype.ToInt32(i32))

	tt.Equal(i64, ztype.ToInt64(str))
	tt.Equal(i64, ztype.ToInt64(i))
	tt.Equal(i64, ztype.ToInt64(i8))
	tt.Equal(i64, ztype.ToInt64(i16))
	tt.Equal(i64, ztype.ToInt64(i32))
	tt.Equal(i64, ztype.ToInt64(i64))
	tt.Equal(i64, ztype.ToInt64(ui8))
	tt.Equal(i64, ztype.ToInt64(ui))
	tt.Equal(i64, ztype.ToInt64(ui16))
	tt.Equal(i64, ztype.ToInt64(ui32))
	tt.Equal(i64, ztype.ToInt64(ui64))
	tt.Equal(i64, ztype.ToInt64(f3))
	tt.Equal(i64, ztype.ToInt64(f6))
	// 无法转换直接换成0
	tt.Equal(ztype.ToInt64(0), ztype.ToInt64(jj))
	tt.Equal(i64, ztype.ToInt64("0x7b"))
	tt.Equal(i64, ztype.ToInt64("0173"))
	tt.Equal(ztype.ToInt64(1), ztype.ToInt64(b))
	tt.Equal(ztype.ToInt64(0), ztype.ToInt64(false))

	tt.Equal(ztype.ToUint(0), ztype.ToUint(ni))
	tt.Equal(ui, ztype.ToUint(str))
	tt.Equal(ui, ztype.ToUint(ui))
	tt.Equal(ui8, ztype.ToUint8(str))
	tt.Equal(ui8, ztype.ToUint8(ui8))
	tt.Equal(ui16, ztype.ToUint16(str))
	tt.Equal(ui16, ztype.ToUint16(ui16))
	tt.Equal(ui32, ztype.ToUint32(str))
	tt.Equal(ui32, ztype.ToUint32(ui32))

	tt.Equal(ui64, ztype.ToUint64(i64))
	tt.Equal(ui64, ztype.ToUint64(str))
	tt.Equal(ui64, ztype.ToUint64(i))
	tt.Equal(ui64, ztype.ToUint64(i8))
	tt.Equal(ui64, ztype.ToUint64(i16))
	tt.Equal(ui64, ztype.ToUint64(i32))
	tt.Equal(ui64, ztype.ToUint64(ui))
	tt.Equal(ui64, ztype.ToUint64(ui8))
	tt.Equal(ui64, ztype.ToUint64(ui16))
	tt.Equal(ui64, ztype.ToUint64(ui32))
	tt.Equal(ui64, ztype.ToUint64(ui64))
	tt.Equal(ui64, ztype.ToUint64(f3))
	tt.Equal(ui64, ztype.ToUint64(f6))
	// 无法转换直接换成0
	tt.Equal(ztype.ToUint64(0), ztype.ToUint64(jj))
	tt.Equal(ui64, ztype.ToUint64("0x7b"))
	tt.Equal(ui64, ztype.ToUint64("0173"))
	tt.Equal(ztype.ToUint64(1), ztype.ToUint64(b))
	tt.Equal(ztype.ToUint64(0), ztype.ToUint64(false))

	tt.Equal(str, ztype.ToString(sst))
	tt.Equal("", ztype.ToString(ni))
	tt.Equal("true", ztype.ToString(b))
	tt.Equal(str, ztype.ToString(str))
	tt.Equal(str, ztype.ToString(i8))
	tt.Equal(str, ztype.ToString(ui))
	tt.Equal(str, ztype.ToString(i))
	tt.Equal(str, ztype.ToString(i8))
	tt.Equal(str, ztype.ToString(i16))
	tt.Equal(str, ztype.ToString(i32))
	tt.Equal(str, ztype.ToString(i64))
	tt.Equal(str, ztype.ToString(ui8))
	tt.Equal(str, ztype.ToString(ui16))
	tt.Equal(str, ztype.ToString(ui32))
	tt.Equal(str, ztype.ToString(ui64))
	tt.Equal(str, ztype.ToString(f6))
	tt.Equal(str, ztype.ToString(f3))
	tt.Equal(str, ztype.ToString(ztype.ToBytes(i)))
	tt.Equal("{\"Name\":\"123\",\"Key\":\"\",\"age\":0}", ztype.ToString(jj))
	tt.Equal(f6, ztype.ToFloat64(i))
	tt.Equal(f6, ztype.ToFloat64(f3))
	tt.Equal(f6, ztype.ToFloat64(f6))
	tt.Equal(ztype.ToFloat64(0), ztype.ToFloat64(ni))

	tt.Equal(f3, ztype.ToFloat32(i))
	tt.Equal(f3, ztype.ToFloat32(f3))
	tt.Equal(f3, ztype.ToFloat32(f6))
	tt.Equal(ztype.ToFloat32(0), ztype.ToFloat32(ni))

	tt.Equal(true, ztype.ToBool(b))
	tt.Equal(true, ztype.ToBool(str))
	tt.Equal(false, ztype.ToBool(ni))

	v := map[string]interface{}{
		"D":  true,
		"E":  12,
		"fs": []string{"1", "a"},
		"gg": map[string]string{"a": "1"},
	}
	var d type2
	tt.NoError(ztype.To(v, &d, func(conver *ztype.Conver) {

	}))
	tt.Log(d)
	tt.Equal(1, d.G["a"])
	tt.EqualTrue(d.D)
	tt.Equal(uint(12), *d.E)
}

func BenchmarkToString1(b *testing.B) {
	s := true
	for i := 0; i < b.N; i++ {
		_ = ztype.ToString(s)
	}
}

func BenchmarkToString2(b *testing.B) {
	s := true
	for i := 0; i < b.N; i++ {
		_ = String(s)
	}
}
func String(val interface{}) string {
	if val == nil {
		return ""
	}

	switch t := val.(type) {
	case bool:
		return strconv.FormatBool(t)
	case int:
		return strconv.FormatInt(int64(t), 10)
	case int8:
		return strconv.FormatInt(int64(t), 10)
	case int16:
		return strconv.FormatInt(int64(t), 10)
	case int32:
		return strconv.FormatInt(int64(t), 10)
	case int64:
		return strconv.FormatInt(t, 10)
	case uint:
		return strconv.FormatUint(uint64(t), 10)
	case uint8:
		return strconv.FormatUint(uint64(t), 10)
	case uint16:
		return strconv.FormatUint(uint64(t), 10)
	case uint32:
		return strconv.FormatUint(uint64(t), 10)
	case uint64:
		return strconv.FormatUint(t, 10)
	case float32:
		return strconv.FormatFloat(float64(t), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case []byte:
		return string(t)
	case string:
		return t
	default:
		return fmt.Sprintf("%v", val)
	}
}

func TestStructToMap(tt *testing.T) {
	e := uint(8)
	t := zls.NewTest(tt)
	v := &type2{
		D: true,
		E: &e,
		F: []string{"f1", "f2"},
		G: map[string]int{"G1": 1, "G2": 2},
		type1: type1{
			A: 1,
			B: "type1",
		},
		S1: type1{
			A: 2,
			B: "S1",
		},
		S2: &type1{
			A: 3,
			B: "Ss",
		},
	}
	r := ztype.ToMap(v)
	t.Log(r, v)

	t.EqualExit(true, r.Get("D").Bool())
	t.EqualExit(8, r.Get("E").Int())
	t.EqualExit(2, r.Get("gg").Get("G2").Int())
	t.EqualExit("2", r.Get("S1").Get("A").String())
	t.EqualExit(r.Get("S1.A").String(), r.Get("S1").Get("A").String())
	t.EqualExit("f2", r.Get("fs").SliceString()[1])
	t.EqualExit(r.Get("fs.1").String(), r.Get("fs").SliceString()[1])
	v2 := []string{"1", "2", "more"}
	r = ztype.ToMap(v2)
	t.Log(v2, r)
	j, err := zjson.Marshal(v2)
	t.EqualNil(err)
	t.EqualExit(`["1","2","more"]`, string(j))

	v3 := "ok"
	r = ztype.ToMap(v3)
	t.Log(v3, r)
	j, err = zjson.Marshal(v3)
	t.EqualNil(err)
	t.EqualExit(`"ok"`, string(j))
}

func TestToTime(t *testing.T) {
	t.Log(ztype.ToTime(1683280800000))
	t.Log(ztype.ToTime(1677670200000))
	t.Log(ztype.ToTime(1658049838))
	t.Log(ztype.ToTime("2022-07-17 17:23:58"))
}

func TestToStruct(t *testing.T) {
	tt := zls.NewTest(t)
	v := map[string]interface{}{
		"D":  true,
		"E":  12,
		"fs": []string{"1", "a"},
	}
	var d type2

	tt.NoError(ztype.ToStruct(v, &d))

	tt.Equal(true, d.D)
	tt.Equal(2, len(d.F))
	t.Log(d)
}
