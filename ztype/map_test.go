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

	err = m.Set("val", "99")
	t.Log(err)
	tt.EqualTrue(err != nil)

	var m2 = &Map{}
	tt.Equal(true, m2.IsEmpty())
	err = m.Delete("no")
	t.Log(err)

	err = m2.Set("val", "99")
	tt.NoError(err)
	tt.Equal("99", m2.Get("val").String())
}

type other struct {
	Sex int
}

type (
	Str string
	Obj struct {
		Name Str `json:"name"`
	}
	u struct {
		Other  *other
		Name   Str                      `json:"name"`
		Region struct{ Country string } `json:"reg"`
		Objs   []Obj
		Key    int
		Status bool
	}
)

var user = &u{
	Name: "n666",
	Key:  9,
	Objs: []Obj{
		{"n1"},
		{"n2"},
	},
	Status: true,
	Region: struct {
		Country string
	}{"中国"},
}

func TestToMap(T *testing.T) {
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
	userMap := Map{
		"Name":   user.Name,
		"Status": user.Status,
		"Key":    user.Key,
	}
	toUserMap := ToMap(user)
	t.Log(user)
	t.Log(userMap)
	t.Log(toUserMap)
	t.Equal(userMap, toUserMap)
}

func TestToMaps(T *testing.T) {
	t := zlsgo.NewTest(T)
	type u struct {
		Other  *other
		Name   string
		Key    int
		Status bool
	}
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
	toSliceMapString := ToMaps(rawData)
	t.Log(toSliceMapString)
	t.Equal(18, toSliceMapString[0].Get("Other").Get("Sex").Int())

	var data = make([]map[string]interface{}, 2)
	data[0] = map[string]interface{}{"name": "hi"}
	data[1] = map[string]interface{}{"name": "golang"}
	toSliceMapString = ToMaps(data)
	t.Equal("hi", toSliceMapString.Index(0).Get("name").String())

	data2 := Maps{{"name": "hi"}, {"name": "golang"}}
	toSliceMapString = ToMaps(data2)
	t.Equal("hi", toSliceMapString.Index(0).Get("name").String())
}
