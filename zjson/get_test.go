package zjson

import (
	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zstring"
	"strings"
	"testing"
	"time"
)

type Demo struct {
	I       int    `json:"i"`
	Quality string `json:"quality"`
}

func TestGet(T *testing.T) {
	t := zlsgo.NewTest(T)
	UnmarshalValidationEnabled(false)
	UnmarshalValidationEnabled(true)
	SetModifiersState(true)
	quality := Get(demo, "quality")
	t.EqualExit("highLevel", quality.String())
	user := Get(demo, "user")
	name := user.Get("name").String()
	other := Get(demo, "other")
	t.Log(other.Array())
	t.EqualExit("暴龙兽", name)
	t.EqualExit(666, Get(demo, "other.1").Int())
	t.Log(Get(demo, "other.1").Type.String())
	t.EqualExit(0, Get(demo, "other.2").Int())
	t.Log(Get(demo, "other.2").Type.String())
	t.EqualExit(0, Get(demo, "bool").Int())
	t.Log(Get(demo, "bool").Type.String())
	t.EqualExit(1, Get(demo, "boolTrue").Int())
	t.EqualExit(0, Get(demo, "time").Int())
	_ = Get(demo, "time").Type.String()
	_ = Get(demo, "timeNull").Type.String()
	t.EqualExit(1.8, Get(demo, "other.2").Float())
	t.EqualExit(66.6, Get(demo, "index\\.key").Float())
	t.EqualExit(uint(666), Get(demo, "other.1").Uint())
	t.EqualExit("666", Get(demo, "other.1").String())
	t.EqualExit(false, Get(demo, "bool").Bool())
	_ = Get(demo, "boolTrue").Type.String()
	t.EqualExit("false", Get(demo, "bool").String())
	t.EqualExit(true, Get(demo, "boolTrue").Bool())
	t.EqualExit(false, Get(demo, "boolTrueNot").Bool())
	t.EqualExit("true", Get(demo, "boolTrue").String())
	timeStr := Get(demo, "time").String()
	t.EqualExit("2019-09-10 13:48:22", timeStr)
	loc, _ := time.LoadLocation("Local")
	tt, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStr, loc)
	t.EqualExit(tt, Get(demo, "time").Time())
	t.EqualExit(true, Get(demo, "user").IsObject())
	t.EqualExit(true, Get(demo, "user").IsObject())
	t.EqualExit(true, Get(demo, "user").Exists())
	t.EqualExit(true, other.IsArray())
	t.EqualExit(Get(demo, "friends.1").String(), Get(demo, "friends").Get("#(name=天女兽)").String())
	t.EqualExit(2, Get(demo, "friends.#").Int())
	t.EqualExit("天女兽", Get(demo, "friends.#(age>1).name").String())
	t.EqualExit("天女兽", Get(demo, "f?iends.1.name").String())
	t.EqualExit("[\"天女兽\"]", Get(demo, "[friends.1.name]").String())
	t.EqualExit(false, Valid("{{}"))
	t.EqualExit(true, Valid(demo))

	ForEachLine(demo+demo, func(line Res) bool {
		return true
	})

	maps := Get(demo, "user").Value().(map[string]interface{})
	for key, value := range maps {
		t.EqualExit("name", key)
		t.EqualExit("暴龙兽", value.(string))
	}

	parseData := Parse(demo)
	t.Log(parseData.Map())

	other.ForEach(func(key, value Res) bool {
		return true
	})

	byteData := zstring.String2Bytes(demo)
	t.EqualExit(true, ValidBytes(byteData))
	t.EqualExit("暴龙兽", GetBytes(byteData, "user.name").String())

	resData := GetMultiple(demo, "user.name", "f?iends.1.name")
	_ = GetMultipleBytes(byteData, "user.name", "f?iends.1.name")
	t.EqualExit("暴龙兽", resData[0].String())
	t.EqualExit("天女兽", resData[1].String())

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
	t.EqualExit(true, ModifierExists("case"))
	t.EqualExit("HIGHLEVEL", Get(demo, "quality|@case:upper|@reverse").String())
	t.Log(Get(demo, "friends").String())
	t.Log(Get(demo, "friends|@reverse|@case:upper").String())
	t.Log(Get(demo, "friends|@format:{\"indent\":\"--\"}").String())

	type Demo struct {
		I       int    `json:"i"`
		Quality string `json:"quality"`
	}
	var demoData Demo
	demoJson := Ugly(zstring.String2Bytes(demo))
	err := Unmarshal(demoJson, &demoData)
	t.Log(err, demoData)

	err = Unmarshal(zstring.String2Bytes(demo), &demoData)
	t.EqualExit(true, err == nil)
	t.Log(err, demoData)

	err = Unmarshal(demo, &demoData)
	t.EqualExit(true, err == nil)
	t.Log(err, demoData)

	err = Unmarshal("demo", &demoData)
	t.EqualExit(true, err != nil)
	t.Log(err, demoData)

	var i struct {
		I int `json:"i"`
	}
	_ = parseData.Unmarshal(&i)
	t.Log(i)
	t.Log(Get(demo, "friends").Type.String())
	t.Log(parseData.Get("@reverse").String())
}

func TestGetFormat(T *testing.T) {
	SetModifiersState(true)
	t := zlsgo.NewTest(T)
	t.Log(Get(demo, "friends|@format:{\"indent\":\"--\"}").String())
}

func BenchmarkGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Get(demo, "i")
	}
}

func BenchmarkGetBytes(b *testing.B) {
	demoByte := []byte(demo)
	for i := 0; i < b.N; i++ {
		_ = GetBytes(demoByte, "i")
	}
}
