package ztype_test

import (
	"reflect"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/ztype"
)

func TestMapTo(t *testing.T) {
	tt := zlsgo.NewTest(t)
	m := ztype.Map{
		"i":     "123",
		"s":     float64(123),
		"b":     "false",
		"z":     "z",
		"int":   "1",
		"uint":  "1",
		"f64":   "64",
		"t":     "2022-12-12",
		"map":   map[string]string{"a": "1", "2": "n"},
		"slice": []string{"1", "2", "3"},
	}

	tt.Equal(reflect.String, reflect.TypeOf(m["i"]).Kind())
	tt.Equal(123, m.GetToInt("i"))
	tt.Equal(reflect.Int, reflect.TypeOf(m["i"]).Kind())

	tt.Equal(0, m.GetToInt("i2"))
	tt.Equal(456, m.GetToInt("i2", 456))
	tt.Equal(uint(456), m.GetToUint("i2"))
	tt.Equal(float64(456), m.GetToFloat64("i2"))

	tt.Equal(reflect.Float64, reflect.TypeOf(m["s"]).Kind())
	tt.Equal(123, m.GetToInt("s"))
	tt.Equal(reflect.Int, reflect.TypeOf(m["s"]).Kind())

	tt.Equal(false, m.GetToBool("b"))
	tt.Equal(true, m.GetToBool("b2", true))

	tt.Equal("123", m.GetToString("s"))
	tt.Equal("666", m.GetToString("s2", "666"))

	tt.Equal([]byte{122}, m.GetToBytes("z"))
	tt.Equal([]byte{1}, m.GetToBytes("z2", []byte{1}))

	tt.Equal(1, m.GetToInt("int"))
	tt.Equal(9, m.GetToInt("int2", 9))

	tt.Equal(float64(64), m.GetToFloat64("f64"))
	tt.Equal(float64(98), m.GetToFloat64("f642", 98))

	t.Log(m.GetToTime("t"))

	tt.Equal(1, m.GetToMap("map").Get("a").Int())

	ss := m.GetToSlice("slice")
	tt.Equal(3, ss.Index(2).Int())

}
