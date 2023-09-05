package ztype

import (
	"reflect"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zreflect"
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
		{"n3"},
		{"n4"},
		{"n5"},
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

	t.Equal(1, ToMap(map[interface{}]interface{}{"name": 1}).Get("name").Int())
	t.Equal(1, ToMap(map[interface{}]string{"name": "1"}).Get("name").Int())
	t.Equal(1, ToMap(map[interface{}]int{"name": 1}).Get("name").Int())
	t.Equal(1, ToMap(map[interface{}]uint{"name": 1}).Get("name").Int())
	t.Equal(1, ToMap(map[interface{}]float64{"name": 1}).Get("name").Int())
	t.Equal(1, ToMap(map[interface{}]bool{"name": true}).Get("name").Int())
	t.Equal(1, ToMap(map[string]interface{}{"name": 1}).Get("name").Int())
	t.Equal(1, ToMap(map[string]string{"name": "1"}).Get("name").Int())
	t.Equal(1, ToMap(map[string]int{"name": 1}).Get("name").Int())
	t.Equal(1, ToMap(map[string]uint{"name": 1}).Get("name").Int())
	t.Equal(1, ToMap(map[string]float64{"name": 1}).Get("name").Int())
	t.Equal(1, ToMap(map[string]bool{"name": true}).Get("name").Int())
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

func TestConvContainTime(t *testing.T) {
	tt := zlsgo.NewTest(t)

	type JsonTime time.Time

	type S struct {
		Date1 JsonTime `z:"date"`
		Date2 time.Time
		Name  string
	}

	now := time.Now()
	v := map[string]interface{}{
		"date":  now,
		"Date2": now,
		"Name":  "123",
	}

	var s S
	isTime := zreflect.TypeOf(time.Time{})
	err := To(v, &s, func(conver *Conver) {
		conver.ConvHook = func(i reflect.Value, o reflect.Type) (reflect.Value, error) {
			t := i.Type()
			if t == isTime && t.ConvertibleTo(o) {
				return i.Convert(o), nil
			}
			return i, nil
		}
	})
	tt.NoError(err)

	tt.Equal(now.Unix(), time.Time(s.Date1).Unix())
	tt.Equal(now.Unix(), s.Date2.Unix())
}

func BenchmarkName(b *testing.B) {

	b.Run("toMapString", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = toMapString(user)
		}
	})

	b.Run("ToMap", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ToMap(user)
		}
	})

	b.Run("toMapString", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = toMapString(user)
			}
		})
	})

	b.Run("ToMap", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = ToMap(user)
			}
		})
	})
}
