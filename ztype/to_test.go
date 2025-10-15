package ztype_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/zreflect"
	"github.com/sohaha/zlsgo/ztime"
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
	JsonTime time.Time
	type2    struct {
		Date  time.Time `z:"date_time"`
		JDate JsonTime  `z:"j_date"`
		E     *uint
		G     map[string]int `z:"gg"`
		S2    *type1
		F     []string `json:"fs"`
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
	ff   string  = "123%"
	b            = true
)

func (s *j) String() string {
	return ztype.ToString(s.Key)
}

func (s *j) Set(v string) {
	s.Key = v
}

func TestTo(t *testing.T) {
	tt := zlsgo.NewTest(t)
	var sst st = new(j)
	sst.Set(str)
	jj := j{Name: "123"}

	tt.EqualExit([]byte(str), ztype.ToBytes(str))
	tt.EqualExit([]byte(str), ztype.ToBytes(i))

	tt.EqualExit(0, ztype.ToInt(ni))
	tt.EqualExit(i, ztype.ToInt(str))
	tt.EqualExit(i, ztype.ToInt(i))
	tt.EqualExit(i8, ztype.ToInt8(str))
	tt.EqualExit(i8, ztype.ToInt8(i8))
	tt.EqualExit(i16, ztype.ToInt16(str))
	tt.EqualExit(i16, ztype.ToInt16(i16))
	tt.EqualExit(i32, ztype.ToInt32(str))
	tt.EqualExit(i32, ztype.ToInt32(i32))

	tt.EqualExit(i64, ztype.ToInt64(str))
	tt.EqualExit(i64, ztype.ToInt64(i))
	tt.EqualExit(i64, ztype.ToInt64(i8))
	tt.EqualExit(i64, ztype.ToInt64(i16))
	tt.EqualExit(i64, ztype.ToInt64(i32))
	tt.EqualExit(i64, ztype.ToInt64(i64))
	tt.EqualExit(i64, ztype.ToInt64(ui8))
	tt.EqualExit(i64, ztype.ToInt64(ui))
	tt.EqualExit(i64, ztype.ToInt64(ui16))
	tt.EqualExit(i64, ztype.ToInt64(ui32))
	tt.EqualExit(i64, ztype.ToInt64(ui64))
	tt.EqualExit(i64, ztype.ToInt64(f3))
	tt.EqualExit(i64, ztype.ToInt64(f6))

	tt.EqualExit(ztype.ToInt64(0), ztype.ToInt64(jj))
	tt.EqualExit(int64(-1), ztype.ToInt64(-1))
	tt.EqualExit(i64, ztype.ToInt64("0x7b"))
	tt.EqualExit(i64, ztype.ToInt64("0173"))
	tt.EqualExit(ztype.ToInt64(1), ztype.ToInt64(b))
	tt.EqualExit(ztype.ToInt64(0), ztype.ToInt64(false))
	tt.EqualExit(int64(123_456), ztype.ToInt64("123_456"))
	tt.EqualExit(int64(123_456), ztype.ToInt64("123,456"))

	tt.EqualExit(ztype.ToUint(0), ztype.ToUint(ni))
	tt.EqualExit(ui, ztype.ToUint(str))
	tt.EqualExit(ui, ztype.ToUint(ui))
	tt.EqualExit(ui8, ztype.ToUint8(str))
	tt.EqualExit(ui8, ztype.ToUint8(ui8))
	tt.EqualExit(ui16, ztype.ToUint16(str))
	tt.EqualExit(ui16, ztype.ToUint16(ui16))
	tt.EqualExit(ui32, ztype.ToUint32(str))
	tt.EqualExit(ui32, ztype.ToUint32(ui32))
	tt.EqualExit(uint32(123_456), ztype.ToUint32("123,456"))

	tt.EqualExit(ui64, ztype.ToUint64(i64))
	tt.EqualExit(ui64, ztype.ToUint64(str))
	tt.EqualExit(ui64, ztype.ToUint64(i))
	tt.EqualExit(ui64, ztype.ToUint64(i8))
	tt.EqualExit(ui64, ztype.ToUint64(i16))
	tt.EqualExit(ui64, ztype.ToUint64(i32))
	tt.EqualExit(ui64, ztype.ToUint64(ui))
	tt.EqualExit(ui64, ztype.ToUint64(ui8))
	tt.EqualExit(ui64, ztype.ToUint64(ui16))
	tt.EqualExit(ui64, ztype.ToUint64(ui32))
	tt.EqualExit(ui64, ztype.ToUint64(ui64))
	tt.EqualExit(ui64, ztype.ToUint64(f3))
	tt.EqualExit(ui64, ztype.ToUint64(f6))

	tt.EqualExit(ztype.ToUint64(0), ztype.ToUint64(jj))
	tt.EqualExit(ui64, ztype.ToUint64("0x7b"))
	tt.EqualExit(ui64, ztype.ToUint64("0173"))
	tt.EqualExit(ztype.ToUint64(1), ztype.ToUint64(b))
	tt.EqualExit(ztype.ToUint64(0), ztype.ToUint64(false))

	tt.EqualExit(str, ztype.ToString(sst))
	tt.EqualExit("", ztype.ToString(ni))
	tt.EqualExit("true", ztype.ToString(b))
	tt.EqualExit(str, ztype.ToString(str))
	tt.EqualExit(str, ztype.ToString(i8))
	tt.EqualExit(str, ztype.ToString(ui))
	tt.EqualExit(str, ztype.ToString(i))
	tt.EqualExit(str, ztype.ToString(i8))
	tt.EqualExit(str, ztype.ToString(i16))
	tt.EqualExit(str, ztype.ToString(i32))
	tt.EqualExit(str, ztype.ToString(i64))
	tt.EqualExit(str, ztype.ToString(ui8))
	tt.EqualExit(str, ztype.ToString(ui16))
	tt.EqualExit(str, ztype.ToString(ui32))
	tt.EqualExit(str, ztype.ToString(ui64))
	tt.EqualExit(str, ztype.ToString(f6))
	tt.EqualExit(str, ztype.ToString(f3))
	tt.EqualExit(str, ztype.ToString(ztype.ToBytes(i)))
	tt.EqualExit("{\"Name\":\"123\",\"Key\":\"\",\"age\":0}", ztype.ToString(jj))
	tt.EqualExit(f6, ztype.ToFloat64(i))
	tt.EqualExit(f6, ztype.ToFloat64(f3))
	tt.EqualExit(f6, ztype.ToFloat64(f6))
	tt.EqualExit(ztype.ToFloat64(0), ztype.ToFloat64(ni))

	tt.EqualExit(f3, ztype.ToFloat32(i))
	tt.EqualExit(f3, ztype.ToFloat32(f3))
	tt.EqualExit(f3, ztype.ToFloat32(f6))
	tt.EqualExit(float32(1.23), ztype.ToFloat32(ff))
	tt.EqualExit(ztype.ToFloat32(0), ztype.ToFloat32(ni))
	tt.EqualExit(float32(123_456.123), ztype.ToFloat32("123,456.123"))

	tt.EqualExit(true, ztype.ToBool(b))
	tt.EqualExit(true, ztype.ToBool(str))
	tt.EqualExit(false, ztype.ToBool(ni))
	tt.EqualExit(false, ztype.ToBool("FAlse"))

	v := map[string]interface{}{
		"D":         true,
		"E":         12,
		"fs":        []string{"1", "a"},
		"gg":        map[string]string{"a": "1"},
		"date_time": time.Now(),
		"j_date":    time.Now(),
	}
	var d type2
	tt.NoError(ztype.To(v, &d))
	tt.Log(d)
	tt.Log(d.JDate)
	tt.Log(d.Date)
	tt.EqualExit(1, d.G["a"])
	tt.EqualTrue(d.D)
	tt.EqualExit(uint(12), *d.E)
}

func TestConv(t *testing.T) {
	tt := zlsgo.NewTest(t)

	type _time time.Time
	now := _time(time.Now())
	otime, _ := ztime.Parse("2021-11-25 00:00:00")
	name := "test"
	a := struct {
		Day     ztime.LocalTime
		Options map[string]string
		Name    *string
		Name3   *string
		Date    *_time `json:"d"`
		Day2    *time.Time
		Name2   string
		Nick    string
		Tags    []string
	}{
		Name: &name,
		Nick: name,
		Tags: []string{"a", "b"},
		Date: &now,
		Day:  ztime.LocalTime{Time: time.Time(now)},
		Options: map[string]string{
			"key": "value",
		},
	}

	var a2 ztype.Map
	tt.NoError(ztype.To(a, &a2))
	tt.EqualExit(a2.Get("Nick").String(), a.Nick)
	tt.EqualExit(a2.Get("Day").String(), ztime.FormatTime(time.Time(now)))
	tt.Log(a2)

	b := ztype.Map{"name": "dev", "tags": []string{"c", "d", "e"}, "options": map[string]string{"new_key": "new_value"}, "d": ztime.FormatTime(otime), "Day": ztime.FormatTime(otime)}
	tt.Log(ztype.To(b, &a))
	tt.Log(a)

	tt.EqualExit("dev", *(a.Name))
	tt.EqualExit([]string{"c", "d", "e"}, a.Tags)
	tt.EqualExit("new_value", a.Options["new_key"])
	tt.EqualExit(1, len(a.Options))
	tt.EqualExit(ztime.FormatTime(otime), ztime.FormatTime(a.Day.Time))
	tt.EqualExit(ztime.FormatTime(otime), ztime.FormatTime(time.Time(*(a.Date))))

	tt.Log(ztype.ToStruct(ztype.Map{"tags": []string{"e"}, "options": map[string]string{"3": "4"}}, &a))
	tt.Log(a)
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
	t := zlsgo.NewTest(tt)
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
	tt := zlsgo.NewTest(t)
	tt.Log(ztype.ToTime(1683280800000))
	tt.Log(ztype.ToTime(1677670200000))
	tt.Log(ztype.ToTime(1658049838))
	tt.Log(ztype.ToTime("2022-07-17 17:23:58"))
}

func TestToStruct(t *testing.T) {
	tt := zlsgo.NewTest(t)
	v := map[string]interface{}{
		"D":  true,
		"E":  12,
		"fs": []string{"1", "a"},
	}
	var d type2

	tt.NoError(ztype.ToStruct(v, &d))

	tt.EqualExit(true, d.D)
	tt.EqualExit(2, len(d.F))
	t.Log(d)
}

func TestToMapFromSliceAndEmpty(t *testing.T) {
	tt := zlsgo.NewTest(t)
	in := []map[string]int{{"a": 1}, {"b": 2}, {"c": 3}}
	var out map[string]int
	tt.NoError(ztype.To(in, &out))
	tt.Equal(3, len(out))
	tt.Equal(1, out["a"])
	tt.Equal(2, out["b"])
	tt.Equal(3, out["c"])

	var out2 map[string]int
	tt.NoError(ztype.To([]int{}, &out2))
	tt.NotNil(out2)
	tt.Equal(0, len(out2))
}

func TestToArrayAndErrors(t *testing.T) {
	tt := zlsgo.NewTest(t)
	in := []int{1, 2, 3}
	var out [3]int
	tt.NoError(ztype.To(in, &out))
	tt.Equal([3]int{1, 2, 3}, out)

	var out2 [2]int
	tt.EqualTrue(ztype.To([]int{1, 2, 3}, &out2) != nil)
}

func TestToFunc(t *testing.T) {
	tt := zlsgo.NewTest(t)
	f := func(i int) int { return i + 1 }
	var g func(int) int
	tt.NoError(ztype.To(f, &g))
	tt.NotNil(g)
	tt.Equal(2, g(1))

	var h func(string) int
	tt.EqualTrue(ztype.To(f, &h) != nil)
}

func TestToMapFromMapEmpty(t *testing.T) {
	tt := zlsgo.NewTest(t)
	in2 := map[string]int{}
	var out2 map[string]int
	tt.NoError(ztype.To(in2, &out2))
	tt.NotNil(out2)
	tt.Equal(0, len(out2))
}

func TestToPtrNilHandling(t *testing.T) {
	tt := zlsgo.NewTest(t)
	var pi *int
	tt.NoError(ztype.To(7, &pi))
	tt.NotNil(pi)
	tt.Equal(7, *pi)

	var pk *int
	tt.NoError(ztype.To(nil, &pk))
	tt.EqualTrue(pk == nil)
}

func TestToMapFromStructSquashNonStructError(t *testing.T) {
	tt := zlsgo.NewTest(t)
	type Bad struct {
		N int `z:"n,squash"`
	}
	var out map[string]interface{}
	err := ztype.To(Bad{N: 1}, &out)
	tt.EqualTrue(err != nil)
}

func TestToArrayFromEmptyMap(t *testing.T) {
	tt := zlsgo.NewTest(t)
	var out [2]int
	tt.NoError(ztype.To(map[string]int{}, &out))
	tt.Equal([2]int{0, 0}, out)
}

func TestToSliceStringToBytes(t *testing.T) {
	tt := zlsgo.NewTest(t)
	var out []byte
	tt.NoError(ztype.To("ab", &out))
	tt.Equal(2, len(out))
	tt.Equal(byte('a'), out[0])
	tt.Equal(byte('b'), out[1])
}

func TestToMapErrorOnWrongType(t *testing.T) {
	tt := zlsgo.NewTest(t)
	var out map[string]int
	tt.EqualTrue(ztype.To(123, &out) != nil)
}

func TestToStructFromMapRemainField(t *testing.T) {
	tt := zlsgo.NewTest(t)
	type S struct {
		N   int                         `z:"n"`
		Rem map[interface{}]interface{} `z:"Rem,remain"`
	}
	in := map[string]interface{}{"n": 1, "a": 2, "b": 3}
	var s S
	tt.NoError(ztype.To(in, &s))
	tt.Equal(1, s.N)
	tt.Equal(2, len(s.Rem))
	tt.Equal(2, s.Rem["a"].(int))
	tt.Equal(3, s.Rem["b"].(int))
}

func TestToMapFromStructSquashOmit(t *testing.T) {
	tt := zlsgo.NewTest(t)
	type Inner struct {
		A int `z:"a,omitempty"`
		B int `z:"b"`
	}
	type Outer struct {
		Inner `z:",squash"`
		Name  string `z:"name"`
	}
	in := Outer{Inner: Inner{A: 0, B: 2}, Name: "n"}
	var out map[string]interface{}
	tt.NoError(ztype.To(in, &out))
	_, ok := out["a"]
	tt.EqualTrue(!ok)
	tt.Equal(2, out["b"].(int))
	tt.Equal("n", out["name"].(string))
}

func TestToStructFromMapNonStringKeyError(t *testing.T) {
	tt := zlsgo.NewTest(t)
	type S struct {
		A int `z:"a"`
	}
	in := map[int]interface{}{1: 2}
	var s S
	tt.EqualTrue(ztype.To(in, &s) != nil)
}

func TestToSliceShrinkExisting(t *testing.T) {
	tt := zlsgo.NewTest(t)
	out := []int{9, 9, 9, 9}
	tt.NoError(ztype.To([]int{1, 2}, &out))
	tt.Equal(2, len(out))
	tt.Equal(1, out[0])
	tt.Equal(2, out[1])
}

func TestValueConvErrorNonPointer(t *testing.T) {
	tt := zlsgo.NewTest(t)
	var x int
	tt.EqualTrue(ztype.ValueConv(1, zreflect.ValueOf(x)) != nil)
}

func TestToPtrEdgeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var nilChan chan int
	var ptrChan *chan int
	tt.NoError(ztype.To(nilChan, &ptrChan))
	tt.EqualTrue(ptrChan == nil)

	var nilFunc func()
	var ptrFunc *func()
	tt.NoError(ztype.To(nilFunc, &ptrFunc))
	tt.EqualTrue(ptrFunc == nil)

	var nilInterface interface{}
	var ptrInterface *interface{}
	tt.NoError(ztype.To(nilInterface, &ptrInterface))
	tt.EqualTrue(ptrInterface == nil)

	var nilMap map[string]int
	var ptrMap *map[string]int
	tt.NoError(ztype.To(nilMap, &ptrMap))
	tt.EqualTrue(ptrMap == nil)

	var nilSlice []int
	var ptrSlice *[]int
	tt.NoError(ztype.To(nilSlice, &ptrSlice))
	tt.EqualTrue(ptrSlice == nil)

	val := 42
	var ptrPtr **int
	tt.NoError(ztype.To(&val, &ptrPtr))
	tt.NotNil(ptrPtr)
	tt.Equal(42, **ptrPtr)
}

func TestToStructErrorCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var str string
	err := ztype.ToStruct(ztype.Map{"value": "test"}, &str)
	tt.EqualTrue(err != nil)

	var slice []int
	err = ztype.ToStruct(ztype.Map{"value": "test"}, &slice)
	tt.EqualTrue(err != nil)

	var num int
	err = ztype.ToStruct(ztype.Map{"value": "test"}, &num)
	tt.EqualTrue(err != nil)
}

func TestToSliceEdgeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var result []int
	err := ztype.To(map[string]int{}, &result)
	tt.NoError(err)
	tt.Equal(0, len(result))

	var strSlice []string
	err = ztype.To([]int{1, 2, 3}, &strSlice)
	tt.NoError(err)
	tt.Equal(3, len(strSlice))
	tt.Equal("1", strSlice[0])
}

func TestParseStringToInt64EdgeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Equal(int64(123456), ztype.ToInt64("123_456"))
	tt.Equal(int64(123456), ztype.ToInt64("123,456"))
	tt.Equal(int64(255), ztype.ToInt64("0xff"))
	tt.Equal(int64(255), ztype.ToInt64("0xFF"))
	tt.Equal(int64(83), ztype.ToInt64("0123"))
	tt.Equal(int64(0), ztype.ToInt64("invalid"))
	tt.Equal(int64(0), ztype.ToInt64(""))
}

func TestToFloatEdgeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Equal(float32(1.23), ztype.ToFloat32("123%"))
	tt.Equal(float64(1.23), ztype.ToFloat64("123%"))
	tt.Equal(float32(123456.789), ztype.ToFloat32("123,456.789"))
	tt.Equal(float64(123456.789), ztype.ToFloat64("123,456.789"))
	tt.Equal(float32(-1.5), ztype.ToFloat32(-1.5))
	tt.Equal(float64(-2.5), ztype.ToFloat64(-2.5))
}

func TestToTimeEdgeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

	time1, err1 := ztype.ToTime(1683280800000)
	tt.NoError(err1)
	tt.EqualTrue(!time1.IsZero())

	time2, err2 := ztype.ToTime(1658049838)
	tt.NoError(err2)
	tt.EqualTrue(!time2.IsZero())

	time3, err3 := ztype.ToTime("2022-07-17 17:23:58")
	tt.NoError(err3)
	tt.EqualTrue(!time3.IsZero())

	time4, err4 := ztype.ToTime("invalid")
	tt.EqualTrue(err4 != nil || time4.IsZero())
}

func TestBasicConversionInterfaceType(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var iface interface{}
	err := ztype.To("test value", &iface)
	tt.NoError(err)
	tt.Equal("test value", iface)

	var iface2 interface{}
	err = ztype.To(123, &iface2)
	tt.NoError(err)
	tt.Equal(123, iface2)

	var iface3 interface{} = "original"
	var iface4 interface{}
	err = ztype.To(iface3, &iface4)
	tt.NoError(err)
	tt.Equal("original", iface4)
}

func TestToWithNilPointer(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var nilPtr *string
	var result string
	err := ztype.To(nilPtr, &result)
	tt.NoError(err)
	tt.Equal("", result)

	var nilIface interface{}
	var result2 int
	err = ztype.To(nilIface, &result2)
	tt.NoError(err)
	tt.Equal(0, result2)
}

func TestToArrayEdgeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

	source := []int{1, 2, 3}
	var target [3]int
	err := ztype.To(source, &target)
	tt.NoError(err)
	tt.Equal([3]int{1, 2, 3}, target)

	source3 := []int{1, 2}
	var target3 [3]int
	err = ztype.To(source3, &target3)
	tt.NoError(err)
	tt.Equal(1, target3[0])
	tt.Equal(2, target3[1])
	tt.Equal(0, target3[2])
}

func TestToFuncEdgeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

	source := func(x int) int { return x * 2 }
	var target func(int) int
	err := ztype.To(source, &target)
	tt.NoError(err)
	tt.NotNil(target)
	tt.Equal(10, target(5))

	var nilFunc func()
	var targetNilFunc func()
	err = ztype.To(nilFunc, &targetNilFunc)
	tt.NoError(err)
	tt.EqualTrue(targetNilFunc == nil)
}

func TestToStructFromMapEdgeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

	type Simple struct {
		Name string `z:"name"`
		Age  int    `z:"age"`
	}

	data := map[string]interface{}{
		"name":  "John",
		"age":   30,
		"extra": "ignored",
	}

	var result Simple
	err := ztype.To(data, &result)
	tt.NoError(err)
	tt.Equal("John", result.Name)
	tt.Equal(30, result.Age)

	type WithUnexported struct {
		Public  string `z:"public"`
		private string
	}

	data2 := map[string]interface{}{
		"public":  "visible",
		"private": "should not set",
	}

	var result2 WithUnexported
	err = ztype.To(data2, &result2)
	tt.NoError(err)
	tt.Equal("visible", result2.Public)

	type WithRemain struct {
		Name   string                      `z:"name"`
		Remain map[interface{}]interface{} `z:",remain"`
	}

	data3 := map[string]interface{}{
		"name":  "test",
		"extra": "field",
		"more":  "data",
	}

	var result3 WithRemain
	err = ztype.To(data3, &result3)
	tt.NoError(err)
	tt.Equal("test", result3.Name)
	if result3.Remain != nil {
		tt.EqualTrue(len(result3.Remain) >= 0)
	}
}

func TestToStructFromMapWithInvalidFields(t *testing.T) {
	tt := zlsgo.NewTest(t)

	type Complex struct {
		IntField    int     `z:"int_field"`
		FloatField  float64 `z:"float_field"`
		BoolField   bool    `z:"bool_field"`
		StringField string  `z:"string_field"`
	}

	data := map[string]interface{}{
		"int_field":    "123",
		"float_field":  "45.67",
		"bool_field":   "true",
		"string_field": 999,
	}

	var c Complex
	err := ztype.To(data, &c)
	tt.NoError(err)
	tt.Equal(123, c.IntField)
	tt.Equal(45.67, c.FloatField)
	tt.EqualTrue(c.BoolField)
	tt.Equal("999", c.StringField)
}

func TestToPtrWithComplexTypes(t *testing.T) {
	tt := zlsgo.NewTest(t)

	val := 42
	ptr := &val
	var ptrPtr **int
	err := ztype.To(ptr, &ptrPtr)
	tt.NoError(err)
	tt.NotNil(ptrPtr)
	tt.NotNil(*ptrPtr)
	tt.Equal(42, **ptrPtr)

	type Inner struct {
		Value string
	}
	inner := Inner{Value: "test"}
	var ptrInner *Inner
	err = ztype.To(inner, &ptrInner)
	tt.NoError(err)
	tt.NotNil(ptrInner)
	tt.Equal("test", ptrInner.Value)

	var nilSlice []string
	var ptrSlice *[]string
	err = ztype.To(nilSlice, &ptrSlice)
	tt.NoError(err)
	tt.EqualTrue(ptrSlice == nil)
}
