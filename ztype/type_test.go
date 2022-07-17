package ztype

import (
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("Map", func(t *testing.T) {
		t.Log(New("123").MapString())
		t.Log(New(`{"name": "test"}`).MapString())
		t.Log(New([]string{"1", "2"}).MapString())
		t.Log(New(map[string]interface{}{"abc": 123}).MapString())
	})

	t.Run("Slice", func(t *testing.T) {
		t.Log(New("123").Slice())
		t.Log(New(`{"name": "test"}`).Slice())
		t.Log(New([]string{"1", "2"}).Slice())
		t.Log(New(map[string]interface{}{"abc": 123}).Slice())
	})
}

func TestNewMap(t *testing.T) {
	m := map[string]interface{}{"a": 1, "b": 2.01, "c": []string{"d", "e", "f", "g", "h"}, "r": map[string]int{"G1": 1, "G2": 2}}
	mt := Map(m)

	for _, v := range []string{"a", "b", "c", "d", "r"} {
		typ := mt.Get(v)
		d := map[string]interface{}{
			"value":   typ.Value(),
			"string":  typ.String(),
			"bool":    typ.Bool(),
			"int":     typ.Int(),
			"int8":    typ.Int8(),
			"int16":   typ.Int16(),
			"int32":   typ.Int32(),
			"int64":   typ.Int64(),
			"uint":    typ.Uint(),
			"uint8":   typ.Uint8(),
			"uint16":  typ.Uint16(),
			"uint32":  typ.Uint32(),
			"uint64":  typ.Uint64(),
			"float32": typ.Float32(),
			"float64": typ.Float64(),
			"map":     typ.MapString(),
			"slice":   typ.Slice(),
		}
		t.Logf("%s %+v", v, d)
	}
}
