package ztype

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestMap(t *testing.T) {
	T := zlsgo.NewTest(t)
	m := make(map[interface{}]interface{})
	m["T"] = "test"
	tMapKeyExists := MapKeyExists("T", m)
	T.Equal(true, tMapKeyExists)
}

func TestMapCopy(t *testing.T) {
	tt := zlsgo.NewTest(t)

	z := Map{"a": 1}
	m := Map{"1": 1, "z": z}
	m2 := m.DeepCopy()
	m3 := m

	tt.Equal(m, m2)
	tt.Equal(m, m3)

	m["1"] = 2
	z["a"] = 2

	tt.EqualTrue(m.Get("z.a").String() != m2.Get("z.a").String())
	t.Log(m, m2, m3)
	tt.EqualTrue(m.Get("z.a").String() == m3.Get("z.a").String())
}

func TestMapNil(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var m Map

	tt.Equal(true, m.IsEmpty())
	err := m.Delete("no")
	t.Log(err)

	tt.NoError(m.Set("val", "99"))
	tt.NoError(m.Set("yes", true))
	tt.NoError(m.Set("", "empty"))
	t.Log(m)

	tt.NoError(m.Delete("yes"))

	tt.Equal("empty", m.Get("").String())
	t.Logf("%+v", m)
}

type other struct {
	Sex int
}

func TestToMapInterface(T *testing.T) {
	t := zlsgo.NewTest(T)
	type u struct {
		Name   string
		Key    int
		Status bool
	}
	user := &u{
		Name:   "666",
		Key:    9,
		Status: true,
	}
	userMap := map[string]interface{}{
		"Name":   user.Name,
		"Status": user.Status,
		"Key":    user.Key,
	}
	toUserMap := ToMapString(user)
	t.Log(user)
	t.Log(userMap)
	t.Log(toUserMap)
	t.Equal(userMap, toUserMap)
}

func TestToMapString(T *testing.T) {
	t := zlsgo.NewTest(T)
	type u struct {
		Name   string
		Key    int
		Status bool
	}
	user := &u{
		Name:   "666",
		Key:    9,
		Status: true,
	}
	userMap := map[string]interface{}{
		"Name":   user.Name,
		"Status": user.Status,
		"Key":    user.Key,
	}
	toUserMap := ToMapString(user)
	t.Log(user)
	t.Log(userMap)
	t.Log(toUserMap)
	t.Equal(userMap, toUserMap)
}

func TestToMapStringDeep(T *testing.T) {
	t := zlsgo.NewTest(T)
	type u struct {
		Other  *other
		Name   string
		Key    int
		Status bool
	}

	user := &u{
		Name:   "666",
		Key:    9,
		Status: true,
		Other: &other{
			Sex: 18,
		},
	}
	userMap := map[string]interface{}{
		"Name":   user.Name,
		"Status": user.Status,
		"Key":    user.Key,
		"Sex":    user.Other.Sex,
	}
	toUserMap := ToMapStringDeep(user)
	t.Log(user)
	t.Log(userMap)
	t.Log(toUserMap)
	t.Equal(userMap, toUserMap)
}

func TestToSliceMapString(T *testing.T) {
	t := zlsgo.NewTest(T)
	type u struct {
		Other  *other
		Name   string
		Key    int
		Status bool
	}
	var data = make([]map[string]interface{}, 2)
	data[0] = map[string]interface{}{"name": "hi"}
	data[1] = map[string]interface{}{"name": "golang"}
	var rawData = make([]u, 2)
	rawData[0] = u{
		Name:   "666",
		Key:    9,
		Status: true,
		Other: &other{
			Sex: 18,
		},
	}
	rawData[1] = u{
		Name:   "666",
		Key:    9,
		Status: true,
	}
	toSliceMapString := ToSliceMapString(rawData)
	t.Log(toSliceMapString)
	t.Equal(18, toSliceMapString[0]["Other"].(*other).Sex)
}
