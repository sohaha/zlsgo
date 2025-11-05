package ztype_test

import (
	"errors"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/ztime"
	"github.com/sohaha/zlsgo/ztype"
)

func TestNew(t *testing.T) {
	tt := zlsgo.NewTest(t)

	v := ztype.New(`{"age": "100"}`)
	v2 := ztype.New(v)
	tt.Equal(v.Get("age").Int(), v2.Get("age").Int())
	tt.Equal(v.Get("age2").Int(12), ztype.New(nil).Int(12))

	t.Run("Map", func(t *testing.T) {
		t.Log(ztype.New("123").Map())
		t.Log(ztype.New(`{"name": "test"}`).Map())
		t.Log(ztype.New([]string{"1", "2"}).Map())
		t.Log(ztype.New(map[string]interface{}{"abc": 123}).Map())
	})

	t.Run("Slice", func(t *testing.T) {
		t.Log(ztype.New("123").SliceValue())
		t.Log(ztype.New(`{"name": "test"}`).Maps())
		t.Log(ztype.New([]string{"1", "2"}).SliceInt())
		t.Log(ztype.New(map[string]interface{}{"abc": 123}).Slice())
	})

	t.Run("Time", func(t *testing.T) {
		t.Log(ztype.New("2022-07-17 17:23:58").Time())
		t.Log(ztype.New(time.Now()).Time())
		t.Log(ztype.New(ztime.Now()).Time())
	})
}

func TestNewMap(t *testing.T) {
	m := map[string]interface{}{"a": 1, "b": 2.01, "c": []string{"d", "e", "f", "g", "h"}, "r": map[string]int{"G1": 1, "G2": 2}}
	mt := ztype.Map(m)

	for _, v := range []string{"a", "b", "c", "d", "r", "_"} {
		typ := mt.Get(v)
		d := map[string]interface{}{
			"value":   typ.Value(),
			"bytes":   typ.Bytes([]byte("_")),
			"string":  typ.String("_"),
			"bool":    typ.Bool(false),
			"int":     typ.Int(1),
			"int8":    typ.Int8(1),
			"int16":   typ.Int16(1),
			"int32":   typ.Int32(1),
			"int64":   typ.Int64(1),
			"uint":    typ.Uint(1),
			"uint8":   typ.Uint8(1),
			"uint16":  typ.Uint16(1),
			"uint32":  typ.Uint32(1),
			"uint64":  typ.Uint64(1),
			"float32": typ.Float32(1),
			"float64": typ.Float64(1),
			"map":     typ.Map(),
			"slice_0": typ.Slice().Index(0).String("_s_"),
		}
		t.Logf("%s %+v", v, d)
	}
}

func TestNewMapKeys(t *testing.T) {
	tt := zlsgo.NewTest(t)

	json := `{"a":1,"b.c":2,"d":{"e":3,"f":4},"g":[5,6],"h":{"i":{"j":"100","k":"101"},"o":["p","q",1,16.8]},"0":"00001"}`
	m := zjson.Parse(json).Map()

	var arr ztype.Maps
	_ = zjson.Unmarshal(`[`+json+`]`, &arr)

	tt.EqualTrue(!arr.IsEmpty())
	tt.Equal(1, arr.Len())
	t.Log(arr.Index(0).Get("no").Exists())

	maps := []ztype.Map{ztype.Map(m), arr.Index(0), map[string]interface{}{"a": 1, "b.c": 2, "d": map[string]interface{}{"e": 3, "f": 4}, "g": []interface{}{5, 6}, "h": map[string]interface{}{"i": map[string]interface{}{"j": "100", "k": "101"}, "o": []interface{}{"p", "q", 1, 16.8}}, "0": "00001"}}
	for _, mt := range maps {
		t.Log(mt.Get("0").Value())
		tt.Equal("00001", mt.Get("0").String())

		t.Log(mt.Get("a").Value())
		tt.Equal(1, mt.Get("a").Int())

		t.Log(mt.Get("b.c").Value())
		tt.EqualTrue(!mt.Get("b.c").Exists())
		tt.Equal(0, mt.Get("b.c").Int())

		t.Log(mt.Get("b\\.c").Value())
		tt.EqualTrue(mt.Get("b\\.c").Exists())
		tt.Equal(2, mt.Get("b\\.c").Int())

		d := mt.Get("d")
		t.Log(d.Value())
		tt.EqualTrue(d.Exists())

		t.Log(d.Get("e").Value())
		tt.Equal(3, d.Get("e").Int())

		t.Log(mt.Get("g").Value())
		tt.Equal("6", mt.Get("g.1").String())

		t.Log(mt.Get("h.i.k").Value())
		tt.Equal("101", mt.Get("h.i.k").String())

		t.Log(mt.Get("h.o.3").Value())
		tt.Equal(16.8, mt.Get("h.o.3").Float64())
	}
}

func TestMapSet(t *testing.T) {
	tt := zlsgo.NewTest(t)

	m := ztype.Map{}

	tt.EqualTrue(m.IsEmpty())
	tt.EqualTrue(!m.Get("a").Exists())
	_ = m.Set("a", 1)
	tt.EqualTrue(m.Get("a").Exists())
	tt.Equal(1, m.Get("a").Int())
	tt.EqualTrue(!m.IsEmpty())

	m2 := ztype.Map{}

	tt.EqualTrue(m2.IsEmpty())
	tt.EqualTrue(!m2.Get("a").Exists())
	_ = m2.Set("a", 1)
	tt.EqualTrue(m2.Get("a").Exists())
	tt.Equal(1, m2.Get("a").Int())
	tt.EqualTrue(!m2.IsEmpty())
}

func TestGetTypeCoverage(t *testing.T) {
	if ztype.GetType(nil) != "nil" {
		t.Fatal("GetType(nil) != nil")
	}
	if ztype.GetType(1) != "int" || ztype.GetType(int8(1)) != "int8" || ztype.GetType(uint(1)) != "uint" {
		t.Fatal("GetType integer branches failed")
	}
	if ztype.GetType(1.0) != "float64" || ztype.GetType(float32(1)) != "float32" {
		t.Fatal("GetType float branches failed")
	}
	if ztype.GetType(true) != "bool" || ztype.GetType("s") != "string" || ztype.GetType([]byte("x")) != "[]byte" {
		t.Fatal("GetType basic types failed")
	}
	typ := ztype.GetType(errors.New("x"))
	if typ == "" {
		t.Fatal("GetType non-basic returned empty string")
	}
}

func TestGetTypeMore(t *testing.T) {
	if ztype.GetType(int32(1)) != "int32" || ztype.GetType(uint32(1)) != "uint32" {
		t.Fatal("GetType int32/uint32 failed")
	}
	if ztype.GetType(uint16(1)) != "uint16" || ztype.GetType(int16(1)) != "int16" {
		t.Fatal("GetType int16/uint16 failed")
	}
	if ztype.GetType(uint8(1)) != "uint8" {
		t.Fatal("GetType uint8 failed")
	}
}

func TestTypeDefaults(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var nilType ztype.Type

	tt.Equal("default", nilType.String("default"))
	tt.Equal([]byte("default"), nilType.Bytes([]byte("default")))
	tt.EqualTrue(nilType.Bool(true))
	tt.Equal(42, nilType.Int(42))
	tt.Equal(int8(42), nilType.Int8(42))
	tt.Equal(int16(42), nilType.Int16(42))
	tt.Equal(int32(42), nilType.Int32(42))
	tt.Equal(int64(42), nilType.Int64(42))
	tt.Equal(uint(42), nilType.Uint(42))
	tt.Equal(uint8(42), nilType.Uint8(42))
	tt.Equal(uint16(42), nilType.Uint16(42))
	tt.Equal(uint32(42), nilType.Uint32(42))
	tt.Equal(uint64(42), nilType.Uint64(42))
	tt.Equal(float32(42.5), nilType.Float32(42.5))
	tt.Equal(42.5, nilType.Float64(42.5))
}

func TestTypeConversionsEdgeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

	type TestStruct struct {
		Name string
	}
	ts := TestStruct{Name: "test"}

	wrapped := ztype.New(ts)
	tt.EqualTrue(wrapped.Exists())
	tt.EqualTrue(wrapped.String() != "")

	outer := ztype.New(wrapped)
	tt.Equal(wrapped.String(), outer.String())

	mapData := map[string]interface{}{
		"name": "test",
		"items": []map[string]interface{}{
			{"id": 1, "name": "first"},
			{"id": 2, "name": "second"},
		},
	}
	mapsType := ztype.New(mapData)
	maps := mapsType.Maps()
	tt.EqualTrue(len(maps) > 0)

	emptyMapType := ztype.New(map[string]interface{}{})
	emptyMaps := emptyMapType.Maps()
	tt.EqualTrue(len(emptyMaps) == 0)

	nonMapType := ztype.New("just a string")
	resultMap := nonMapType.Map()
	tt.EqualTrue(len(resultMap) > 0)
}

func TestTypeSliceMethods(t *testing.T) {
	tt := zlsgo.NewTest(t)

	sliceData := []int{1, 2, 3, 4, 5}
	sliceType := ztype.New(sliceData)

	noConvSlice := sliceType.Slice(true)
	tt.Equal(len(sliceData), len(noConvSlice.Value()))

	tt.Equal([]int{1, 2, 3, 4, 5}, sliceType.SliceInt())
	tt.Equal([]string{"1", "2", "3", "4", "5"}, sliceType.SliceString())
	tt.Equal(5, len(sliceType.SliceValue()))

	singleValue := ztype.New(42)
	singleSlice := singleValue.Slice()
	tt.Equal(1, len(singleSlice.Value()))
	tt.Equal(42, singleSlice.Int()[0])

	emptyType := ztype.New(nil)
	emptySlice := emptyType.Slice()
	tt.Equal(0, len(emptySlice.Value()))

	stringType := ztype.New("a,b,c")
	stringSlice := stringType.Slice()
	tt.Equal(1, len(stringSlice.Value()))
}

func TestTypeGetNestedPaths(t *testing.T) {
	tt := zlsgo.NewTest(t)

	nestedData := map[string]interface{}{
		"user": map[string]interface{}{
			"profile": map[string]interface{}{
				"name": "John",
				"age":  30,
			},
			"settings": map[string]interface{}{
				"theme": "dark",
			},
		},
		"items": []interface{}{
			map[string]interface{}{"id": 1, "name": "first"},
			map[string]interface{}{"id": 2, "name": "second"},
		},
	}

	nestedType := ztype.New(nestedData)

	tt.Equal("John", nestedType.Get("user.profile.name").String())
	tt.Equal(30, nestedType.Get("user.profile.age").Int())
	tt.Equal("dark", nestedType.Get("user.settings.theme").String())

	tt.Equal("first", nestedType.Get("items.0.name").String())
	tt.Equal(2, nestedType.Get("items.1.id").Int())

	tt.EqualTrue(!nestedType.Get("user.nonexistent").Exists())
	tt.EqualTrue(!nestedType.Get("user.profile.nonexistent").Exists())
	tt.EqualTrue(!nestedType.Get("items.5.name").Exists())

	nilType := ztype.New(nil)
	tt.EqualTrue(!nilType.Get("any.path").Exists())
}

func TestTypeNumericConversions(t *testing.T) {
	tt := zlsgo.NewTest(t)

	stringInt := ztype.New("123")
	tt.Equal(123, stringInt.Int())
	tt.Equal(int64(123), stringInt.Int64())
	tt.Equal(uint(123), stringInt.Uint())
	tt.Equal(123.0, stringInt.Float64())

	stringFloat := ztype.New("123.456")
	tt.Equal(123, stringFloat.Int())
	tt.Equal(123.456, stringFloat.Float64())
	tt.Equal(float32(123.456), stringFloat.Float32())

	trueType := ztype.New(true)
	tt.Equal(1, trueType.Int())
	tt.Equal(uint(1), trueType.Uint())
	tt.Equal(0.0, trueType.Float64())

	falseType := ztype.New(false)
	tt.Equal(0, falseType.Int())
	tt.Equal(uint(0), falseType.Uint())
	tt.Equal(0.0, falseType.Float64())

	bigNumber := ztype.New("9223372036854775806")
	tt.EqualTrue(bigNumber.Int64() > 0)
	tt.Equal(int64(9223372036854775806), bigNumber.Int64())
}

func TestTypeEdgeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

	type EmptyStruct struct{}
	emptyType := ztype.New(EmptyStruct{})
	tt.EqualTrue(emptyType.Exists())
	tt.EqualTrue(emptyType.String() != "")

	var nilInterface interface{}
	nilInterfaceType := ztype.New(nilInterface)
	tt.EqualTrue(!nilInterfaceType.Exists())

	fnType := ztype.New(func() {})
	tt.EqualTrue(fnType.Exists())
	tt.EqualTrue(fnType.String() == "")

	ch := make(chan int)
	chType := ztype.New(ch)
	tt.EqualTrue(chType.Exists())

	value := 42
	ptrType := ztype.New(&value)
	tt.Equal(42, ptrType.Int())

	var nilPtr *int
	nilPtrType := ztype.New(nilPtr)
	tt.EqualTrue(nilPtrType.Exists())
}

func TestTypeByteConversion(t *testing.T) {
	tt := zlsgo.NewTest(t)

	stringType := ztype.New("test string")
	bytes := stringType.Bytes()
	tt.Equal("test string", string(bytes))

	var nilType ztype.Type
	tt.Equal([]byte("default"), nilType.Bytes([]byte("default")))

	originalBytes := []byte("original")
	bytesType := ztype.New(originalBytes)
	resultBytes := bytesType.Bytes()
	tt.Equal(originalBytes, resultBytes)

	numberType := ztype.New(123)
	numberBytes := numberType.Bytes()
	tt.Equal("123", string(numberBytes))
}
