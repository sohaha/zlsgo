//go:build go1.18
// +build go1.18

package zarray

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/ztype"
)

func TestKeys(t *testing.T) {
	tt := zlsgo.NewTest(t)
	tt.Equal(3, len(Keys(map[int]int{1: 1, 2: 2, 3: 3})))
	tt.Equal(3, len(Keys(map[int]interface{}{1: 1, 2: "2", 3: 3})))
}

func TestValues(t *testing.T) {
	tt := zlsgo.NewTest(t)
	tt.Equal(3, len(Values(map[int]int{1: 1, 2: 2, 3: 3})))
	tt.Equal(3, len(Values(map[int]interface{}{1: 1, 2: "2", 3: 3})))
}

func TestGroupMap(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Run("base", func(tt *zlsgo.TestUtil) {
		value := ztype.Maps{
			{"id": 1, "name": "test"},
			{"id": 2, "name": "test2"},
		}

		data, _ := IndexMap(value, func(v ztype.Map) (string, ztype.Map) {
			id := v.Get("id").String()
			v.Delete("id")
			return "id_" + id, v
		})
		tt.Log(data)
		tt.Equal(2, len(data))
		tt.Equal("test", data["id_1"]["name"])
		tt.Equal("test2", data["id_2"]["name"])
	})

	tt.Run("shallowCopy", func(tt *zlsgo.TestUtil) {
		data1 := map[string]string{"name": "test", "age": "18"}
		value := ztype.Maps{
			{"id": 1, "data": data1},
			{"id": 2, "data": map[string]string{"name": "test2", "age": "19"}},
		}

		tt.Log(data1, value[0])
		data, _ := IndexMap(value, func(v ztype.Map) (string, ztype.Map) {
			id := v.Get("id").String()
			v.Delete("id")
			return "id_" + id, v
		})
		tt.Log(data1, value[0], data)
		data1["name"] = "test_new"
		data1["age"] = "19"
		tt.Log(data1, value[0], data)

		tt.Log(data["id_1"].Get("data"))
		tt.Equal(2, len(data))
		tt.Equal("test_new", data["id_1"].Get("data").Get("name").String())
		tt.Equal("test2", data["id_2"].Get("data").Get("name").String())
	})

	tt.Run("deepCopy", func(tt *zlsgo.TestUtil) {
		data1 := map[string]string{"name": "test", "age": "18"}
		value := ztype.Maps{
			{"id": 1, "data": data1},
			{"id": 2, "data": map[string]string{"name": "test2", "age": "19"}},
		}

		tt.Log(data1, value[0])
		data, _ := IndexMap(value, func(v ztype.Map) (string, ztype.Map) {
			data := v.DeepCopy()
			id := data.Get("id").String()
			data.Delete("id")
			return "id_" + id, data
		})
		tt.Log(data1, value[0], data)
		data1["name"] = "test_new"
		data1["age"] = "19"
		tt.Log(data1, value[0], data)

		tt.Log(data["id_1"].Get("data"))
		tt.Equal(2, len(data))
		tt.Equal("test", data["id_1"].Get("data").Get("name").String())
		tt.Equal("test2", data["id_2"].Get("data").Get("name").String())
	})

	tt.Run("error", func(tt *zlsgo.TestUtil) {
		value := ztype.Maps{
			{"id": 1, "name": "test"},
			{"id": 2, "name": "test2"},
		}

		_, err := IndexMap(value, func(v ztype.Map) (string, ztype.Map) {
			return "id", v
		})
		tt.NotNil(err)
	})
}

func TestFlatMap(t *testing.T) {
	tt := zlsgo.NewTest(t)

	value := map[string]ztype.Map{
		"1": {"name": "test"},
		"2": {"name": "test2"},
	}

	data := FlatMap(value, func(key string, value ztype.Map) ztype.Map {
		value["id"] = key
		return value
	})
	tt.Log(data)
	tt.Equal(2, len(data))

	expected := map[string]map[string]string{
		"1": {"id": "1", "name": "test"},
		"2": {"id": "2", "name": "test2"},
	}

	found := make(map[string]bool)
	for _, item := range data {
		id := item.Get("id").String()
		name := item.Get("name").String()

		exp, exists := expected[id]
		tt.Equal(true, exists)
		if !exists {
			continue
		}

		tt.Equal(exp["name"], name)

		found[id] = true
	}

	tt.Equal(len(expected), len(found))
}
