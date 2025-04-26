package zjson

import (
	"strings"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztype"
)

type Demo struct {
	Quality string `json:"quality"`
	User    struct {
		Name string `json:"name"`
	} `json:"user"`
	Children []string `json:"children"`
	Friends  []struct {
		Name string `json:"name"`
	} `json:"friends"`
	I int `json:"i"`
	F float64
}

type SS struct {
	Name     string   `json:"name"`
	Gg       GG       `json:"g"`
	To       []string `json:"t"`
	IDs      []AA     `json:"ids"`
	Property struct {
		Name string `json:"n"`
		Key  float64
	} `json:"p"`
	Abc int
	ID  int `json:"id"`
	Pid uint
	To2 int `json:"t2"`
}

type GG struct {
	Info string
	P    []AA `json:"p"`
}

type AA struct {
	Name string
	Gg   GG  `json:"g"`
	ID   int `json:"id"`
}

func TestGet2(t *testing.T) {
	Parse(`{"a":null}`).Get("a").ForEach(func(key, value *Res) bool {
		t.Log(key, value)
		t.Fail()
		return true
	})
	get := Parse(`{"a":{"b":"a123",}`).Get("a")
	get.ForEach(func(key, value *Res) bool {
		t.Log(key, value)
		return true
	})
	t.Log(get.str)
	t.Log(get.raw)
	t.Log(get.Get("b"))
}

func TestGet(t *testing.T) {
	tt := zlsgo.NewTest(t)
	SetModifiersState(true)
	quality := Get(demo, "quality")
	tt.EqualExit("highLevel", quality.String())
	user := Get(demo, "user")
	name := user.Get("name").String()
	other := Get(demo, "other")
	t.Log(other.Array(), other.Raw())
	tt.EqualExit("暴龙兽", name)
	tt.EqualExit(name, string(user.Get("name").Bytes()))
	tt.EqualExit("-999", Get(demo, "ii").String())
	tt.EqualExit(666, Get(demo, "other.1").Int())
	tt.Log(Get(demo, "other.1").typ.String())
	tt.EqualExit(0, Get(demo, "other.2").Int())
	tt.Log(Get(demo, "other.2").typ.String())
	tt.EqualExit(0, Get(demo, "bool").Int())
	tt.Log(Get(demo, "bool").typ.String())
	tt.EqualExit(1, Get(demo, "boolTrue").Int())
	tt.EqualExit(int8(1), Get(demo, "boolTrue").Int8())
	tt.EqualExit(int16(1), Get(demo, "boolTrue").Int16())
	tt.EqualExit(int32(1), Get(demo, "boolTrue").Int32())
	tt.EqualExit(int64(1), Get(demo, "boolTrue").Int64())

	tt.EqualExit(0, Get(demo, "time").Int())
	_ = Get(demo, "time").typ.String()
	_ = Get(demo, "timeNull").typ.String()
	tt.EqualExit(1.8, Get(demo, "other.2").Float())
	tt.EqualExit(66.6, Get(demo, "index\\.key").Float())

	tt.EqualExit(uint(666), Get(demo, "other.1").Uint())
	tt.EqualExit(uint(0), Get(demo, "time").Uint())
	tt.EqualExit(uint(1), Get(demo, "f").Uint())
	tt.EqualExit(uint8(1), Get(demo, "f").Uint8())
	tt.EqualExit(uint16(1), Get(demo, "f").Uint16())
	tt.EqualExit(uint32(1), Get(demo, "f").Uint32())
	tt.EqualExit(uint64(1), Get(demo, "f").Uint64())
	tt.EqualExit(uint(0), Get(demo, "user").Uint())
	tt.EqualExit(uint(1), Get(demo, "boolTrue").Uint())

	tt.EqualExit("666", Get(demo, "other.1").String())
	tt.EqualExit(false, Get(demo, "bool").Bool())
	_ = Get(demo, "boolTrue").typ.String()
	tt.EqualExit("false", Get(demo, "bool").String())
	tt.EqualExit(true, Get(demo, "boolTrue").Bool())
	tt.EqualExit(false, Get(demo, "boolTrueNot").Bool())
	tt.EqualExit("true", Get(demo, "boolTrue").String())
	t.Log(Get(demo, "time").Bool(), Get(demo, "time").String())
	tt.EqualExit(false, Get(demo, "time").Bool())
	tt.EqualExit(true, Get(demo, "i").Bool())
	timeStr := Get(demo, "time").String()
	tt.EqualExit("2019-09-10 13:48:22", timeStr)
	loc, _ := time.LoadLocation("Local")
	ttime, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStr, loc)
	tt.EqualExit(ttime, Get(demo, "time").Time())
	tt.EqualExit(true, Get(demo, "user").IsObject())
	tt.EqualExit(true, Get(demo, "user").IsObject())
	tt.EqualExit(true, Get(demo, "user").Exists())
	tt.EqualExit("暴龙兽", Get(demo, "user").Map().Get("name").String())
	tt.EqualExit("天使兽", Get(demo, "friends").Maps().First().Get("name").String())
	tt.EqualExit(Get(demo, "friends").Maps().First().Get("name").String(), Get(demo, "friends").Maps().Index(0).Get("name").String())
	tt.EqualExit(true, other.IsArray())
	tt.EqualExit(Get(demo, "friends.1").String(), Get(demo, "friends").Get("#(name=天女兽)").String())
	tt.EqualExit(2, Get(demo, "friends.#").Int())
	tt.EqualExit("天女兽", Get(demo, "friends.#(age>1).name").String())
	tt.EqualExit("天女兽", Get(demo, "f?iends.1.name").String())
	tt.EqualExit("[\"天女兽\"]", Get(demo, "[friends.1.name]").String())
	tt.EqualExit(false, Valid("{{}"))
	tt.EqualExit(true, Valid(demo))
	tt.EqualExit("阿古兽", Get(demo, "children").SliceString()[0])
	tt.EqualExit(0, Get(demo, "children").SliceInt()[0])

	ForEachLine(demo+demo, func(line *Res) bool {
		return true
	})

	maps := Get(demo, "user").Value().(map[string]interface{})
	for key, value := range maps {
		tt.EqualExit("name", key)
		tt.EqualExit("暴龙兽", value.(string))
	}

	parseData := Parse(demo)
	t.Log(parseData.MapRes())
	t.Log(parseData.MapKeys("children"))

	other.ForEach(func(key, value *Res) bool {
		return true
	})

	Parse(`{"a":null}`).Get("a").ForEach(func(key, value *Res) bool {
		t.Log(key, value)
		t.Fail()
		return true
	})
	Parse(`{"a":"a123"}`).Get("a").ForEach(func(key, value *Res) bool {
		t.Log(key, value)
		return true
	})

	byteData := zstring.String2Bytes(demo)
	tt.EqualTrue(ValidBytes(byteData))
	tt.EqualExit("暴龙兽", GetBytes(byteData, "user.name").String())

	resData := GetMultiple(demo, "user.name", "f?iends.1.name")
	_ = GetMultipleBytes(byteData, "user.name", "f?iends.1.name")
	tt.EqualExit("暴龙兽", resData[0].String())
	tt.EqualExit("天女兽", resData[1].String())

	modifierFn := func(json, arg string) string {
		if arg == "upper" {
			return strings.ToUpper(json)
		}
		if arg == "lower" {
			return strings.ToLower(json)
		}
		return json
	}
	AddModifier("case", modifierFn)
	tt.EqualExit(true, ModifierExists("case"))
	tt.EqualExit("HIGHLEVEL", Get(demo, "quality|@case:upper|@reverse").String())
	t.Log(Get(demo, "friends").String())
	t.Log(Get(demo, "friends|@reverse|@case:upper").String())
	t.Log(Get(demo, "friends|@format:{\"indent\":\"--\"}").String())

	type Demo struct {
		Quality string `json:"quality"`
		I       int    `json:"i"`
	}
	var demoData Demo
	demoJson := Ugly(zstring.String2Bytes(demo))
	err := Unmarshal(demoJson, &demoData)
	t.Log(err, demoData, string(demoJson))

	err = Unmarshal(zstring.String2Bytes(demo), &demoData)
	tt.EqualExit(true, err == nil)
	tt.Log(err, demoData)

	err = Unmarshal(demo, &demoData)
	tt.EqualExit(true, err == nil)
	tt.Log(err, demoData)

	err = Unmarshal("demo", &demoData)
	tt.EqualExit(true, err != nil)
	tt.Log(err, demoData)

	var i struct {
		I int `json:"i"`
	}
	_ = parseData.Unmarshal(&i)
	tt.Log(i)
	tt.Log(Get(demo, "friends").typ.String())
	tt.Log(parseData.Get("@reverse").String())
}

func TestForEach(t *testing.T) {
	tt := zlsgo.NewTest(t)
	arr := Parse(`{"names":[{"name":1},{"name":2}],"values":[3,4]}`)
	arr.ForEach(func(key, value *Res) bool {
		tt.Log("key:", key, "value:", value.String())
		return true
	})

	i := 0
	arr.Get("names").ForEach(func(key, value *Res) bool {
		tt.Log("key:", key.Int(), "value:", value.String())
		tt.Equal(i, key.Int())
		i++
		return true
	})

	Parse(`[{"fen": 63.12, "date": "2023-08-24"}]`).ForEach(func(key, value *Res) bool {
		tt.Log(key, value)
		tt.Equal(63.12, value.Get("fen").Float())
		return true
	})
}

func TestUnmarshal(t *testing.T) {
	tt := zlsgo.NewTest(t)
	json := `{"id":666,"Pid":100,"name":"HelloWorld","g":{"Info":"基础"},"ids":[{"id":1,"Name":"用户1","g":{"Info":"详情","p":[{"Name":"n1","id":1},{"id":2}]}}]}`
	var s SS
	err := Unmarshal(json, &s)
	tt.NoError(err)
	tt.Logf("%+v", s)
	tt.Equal("用户1", s.IDs[0].Name)
	tt.Equal("n1", s.IDs[0].Gg.P[0].Name)
}

func TestUnmarshal2(t *testing.T) {
	tt := zlsgo.NewTest(t)
	json := `{"u1":[{"name":"HH","id":1},{"name":"HBB","id":2}]}`
	var m map[string][]map[string]any
	err := Unmarshal(json, &m)
	tt.NoError(err)
	tt.Logf("%+v", m)
	tt.Equal("HH", m["u1"][0]["name"])
	tt.Equal(2.0, m["u1"][1]["id"])

	json = `{"u2":{"u3":1}}`
	var m2 map[string]map[string]any
	err = Unmarshal(json, &m2)
	tt.NoError(err)
	tt.Logf("%+v", m2)
	tt.Equal(1.0, m2["u2"]["u3"])

	json = `{"u4":2}`
	var m3 map[string]int
	err = Unmarshal(json, &m3)
	tt.NoError(err)
	tt.Logf("%+v", m3)
	tt.Equal(2, m3["u4"])
}

func TestEditJson(t *testing.T) {
	tt := zlsgo.NewTest(t)
	j := Parse(demo)

	name := j.Get("user.name").String()
	_ = j.Delete("user.name")
	nName := j.Get("user.name").String()
	tt.Equal("", nName)
	tt.Equal("暴龙兽", name)

	c1 := j.Get("children.0").String()
	_ = j.Delete("children.0")
	_ = j.Delete("children.0")
	nc1 := j.Get("children.0").String()
	tt.Equal("阿古兽", c1)
	tt.Equal("机器暴龙兽", nc1)

	_ = j.Set("new_path.0", "仙人掌兽")
	_ = j.Set("new_path.1", "花仙兽")

	t.Log(j.Get("friends.0").Map())

	t.Log(string(Ugly([]byte(j.String()))))
}

func TestGetFormat(t *testing.T) {
	SetModifiersState(true)
	t.Log(Get(demo, "friends|@format:{\"indent\":\"--\"}").String())
}

func TestModifiers(t *testing.T) {
	SetModifiersState(true)
	t.Log(Get(demo, "friends").String())
	t.Log(Get(demo, "friends|@reverse").String())
	t.Log(Get(demo, "friends|@ugly").String())
	t.Log(Get(demo, "friends|@format").String())
}

func TestType(t *testing.T) {
	tt := zlsgo.NewTest(t)
	tt.EqualExit(float64(1), Get(`{"a":true}`, "a").Float())
	tt.EqualExit(float64(0), Get(`{"a":false}`, "a").Float())
	tt.EqualExit(ztype.Map{}, Get(`{}`, "a").Map())
	tt.EqualExit(ztype.Maps{}, Get(`{}`, "a").Maps())
	tt.EqualExit([]*Res{}, Get(`{}`, "a").Array())
}

func TestDefault(t *testing.T) {
	tt := zlsgo.NewTest(t)
	notExists := Get(demo, "notExists")
	tt.EqualExit(false, notExists.Exists())
	tt.EqualExit("default", notExists.String("default"))
	tt.EqualExit("", notExists.String())

	tt.EqualExit(false, notExists.Bool())
	tt.EqualExit(true, notExists.Bool(true))

	tt.EqualExit(0, notExists.Int())
	tt.EqualExit(1, notExists.Int(1))
	tt.EqualExit(int8(0), notExists.Int8())
	tt.EqualExit(int8(1), notExists.Int8(1))
	tt.EqualExit(int16(0), notExists.Int16())
	tt.EqualExit(int16(1), notExists.Int16(1))
	tt.EqualExit(int32(0), notExists.Int32())
	tt.EqualExit(int32(1), notExists.Int32(1))
	tt.EqualExit(int64(0), notExists.Int64())
	tt.EqualExit(int64(1), notExists.Int64(1))

	tt.EqualExit(float64(0), notExists.Float())
	tt.EqualExit(float64(1), notExists.Float(1.0))
	tt.EqualExit(float64(0), notExists.Float64())
	tt.EqualExit(float64(1), notExists.Float64(1.0))
	tt.EqualExit(float32(0), notExists.Float32())
	tt.EqualExit(float32(1), notExists.Float32(1.0))

	tt.EqualExit(uint(0), notExists.Uint())
	tt.EqualExit(uint(1), notExists.Uint(1))
	tt.EqualExit(uint8(0), notExists.Uint8())
	tt.EqualExit(uint8(1), notExists.Uint8(1))
	tt.EqualExit(uint16(0), notExists.Uint16())
	tt.EqualExit(uint16(1), notExists.Uint16(1))
	tt.EqualExit(uint32(0), notExists.Uint32())
	tt.EqualExit(uint32(1), notExists.Uint32(1))
	tt.EqualExit(uint64(0), notExists.Uint64())
	tt.EqualExit(uint64(1), notExists.Uint64(1))
}

func BenchmarkGet(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Get(demo, "i")
	}
}

func BenchmarkGetBytes(b *testing.B) {
	demoByte := []byte(demo)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetBytes(demoByte, "i")
	}
}

func BenchmarkGetBig(b *testing.B) {
	json := getBigJSON()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Get(json, "i")
	}
}

func BenchmarkGetBigBytes(b *testing.B) {
	json := zstring.String2Bytes(getBigJSON())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetBytes(json, "i")
	}
}
