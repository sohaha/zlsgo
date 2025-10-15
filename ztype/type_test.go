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

	var m2 = ztype.Map{}

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
	if ztype.GetType(int32(1)) != "int32" || ztype.GetType(uint32(1)) != "uint32" { t.Fatal("GetType int32/uint32 failed") }
	if ztype.GetType(uint16(1)) != "uint16" || ztype.GetType(int16(1)) != "int16" { t.Fatal("GetType int16/uint16 failed") }
	if ztype.GetType(uint8(1)) != "uint8" { t.Fatal("GetType uint8 failed") }
}
