package zjson

import (
	"strings"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zstring"
)

type Demo struct {
	I        int `json:"i"`
	F        float64
	Children []string `json:"children"`
	Quality  string   `json:"quality"`
	User     struct {
		Name string `json:"name"`
	} `json:"user"`
	Friends []struct {
		Name string `json:"name"`
	} `json:"friends"`
}

func TestGet(t *testing.T) {
	tt := zlsgo.NewTest(t)
	SetModifiersState(true)
	quality := Get(demo, "quality")
	tt.EqualExit("highLevel", quality.String())
	user := Get(demo, "user")
	name := user.Get("name").String()
	other := Get(demo, "other")
	tt.Log(other.Array())
	tt.EqualExit("暴龙兽", name)
	tt.EqualExit(666, Get(demo, "other.1").Int())
	tt.Log(Get(demo, "other.1").Type.String())
	tt.EqualExit(0, Get(demo, "other.2").Int())
	tt.Log(Get(demo, "other.2").Type.String())
	tt.EqualExit(0, Get(demo, "bool").Int())
	tt.Log(Get(demo, "bool").Type.String())
	tt.EqualExit(1, Get(demo, "boolTrue").Int())
	tt.EqualExit(0, Get(demo, "time").Int())
	_ = Get(demo, "time").Type.String()
	_ = Get(demo, "timeNull").Type.String()
	tt.EqualExit(1.8, Get(demo, "other.2").Float())
	tt.EqualExit(66.6, Get(demo, "index\\.key").Float())

	tt.EqualExit(uint(666), Get(demo, "other.1").Uint())
	tt.EqualExit(uint(0), Get(demo, "time").Uint())
	tt.EqualExit(uint(1), Get(demo, "f").Uint())
	tt.EqualExit(uint(0), Get(demo, "user").Uint())
	tt.EqualExit(uint(1), Get(demo, "boolTrue").Uint())

	tt.EqualExit("666", Get(demo, "other.1").String())
	tt.EqualExit(false, Get(demo, "bool").Bool())
	_ = Get(demo, "boolTrue").Type.String()
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
	tt.EqualExit(true, other.IsArray())
	tt.EqualExit(Get(demo, "friends.1").String(), Get(demo, "friends").Get("#(name=天女兽)").String())
	tt.EqualExit(2, Get(demo, "friends.#").Int())
	tt.EqualExit("天女兽", Get(demo, "friends.#(age>1).name").String())
	tt.EqualExit("天女兽", Get(demo, "f?iends.1.name").String())
	tt.EqualExit("[\"天女兽\"]", Get(demo, "[friends.1.name]").String())
	tt.EqualExit(false, Valid("{{}"))
	tt.EqualExit(true, Valid(demo))

	ForEachLine(demo+demo, func(line Res) bool {
		return true
	})

	maps := Get(demo, "user").Value().(map[string]interface{})
	for key, value := range maps {
		tt.EqualExit("name", key)
		tt.EqualExit("暴龙兽", value.(string))
	}

	parseData := Parse(demo)
	tt.Log(parseData.Map())
	tt.Log(parseData.MapKeys())

	other.ForEach(func(key, value Res) bool {
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
	tt.Log(Get(demo, "friends").String())
	tt.Log(Get(demo, "friends|@reverse|@case:upper").String())
	tt.Log(Get(demo, "friends|@format:{\"indent\":\"--\"}").String())

	type Demo struct {
		I       int    `json:"i"`
		Quality string `json:"quality"`
	}
	var demoData Demo
	demoJson := Ugly(zstring.String2Bytes(demo))
	err := Unmarshal(demoJson, &demoData)
	t.Log(err, demoData)

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
	tt.Log(Get(demo, "friends").Type.String())
	tt.Log(parseData.Get("@reverse").String())
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
