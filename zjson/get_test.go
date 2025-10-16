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

type UnescapeTestCase struct {
	input    string
	expected string
}

func Test_unescape(t *testing.T) {
	tt := zlsgo.NewTest(t)

    tests := []zlsgo.TestCase{
		{Name: "BasicASCII", Data: UnescapeTestCase{"hello", "hello"}},
		{Name: "SimpleEscape", Data: UnescapeTestCase{"\\\"", "\""}},
		{Name: "UnicodeBasic", Data: UnescapeTestCase{"\\u0041", "A"}},
		{Name: "UnicodeChinese", Data: UnescapeTestCase{"\\u4F60\\u597D", "你好"}},
		{Name: "SurrogatePair", Data: UnescapeTestCase{"\\uD83D\\uDE00", "😀"}},
		{Name: "IncompleteSurrogate", Data: UnescapeTestCase{"\\uD83D", "\uFFFD"}},
		{Name: "InvalidUnicode", Data: UnescapeTestCase{"\\uXXXX", "\u0000"}},
		{Name: "MixedContent", Data: UnescapeTestCase{"Hello\\u0020World\\t!", "Hello World\t!"}},
		{Name: "ControlCharacter", Data: UnescapeTestCase{"\x19", ""}},
		{Name: "InvalidEscape", Data: UnescapeTestCase{"\\x", ""}},
		{Name: "UnicodeInsufficientChars", Data: UnescapeTestCase{"\\u12", ""}},
		{Name: "Backspace", Data: UnescapeTestCase{"\\b", "\b"}},
		{Name: "FormFeed", Data: UnescapeTestCase{"\\f", "\f"}},
		{Name: "CarriageReturn", Data: UnescapeTestCase{"\\r", "\r"}},
		{Name: "Tab", Data: UnescapeTestCase{"\\t", "\t"}},
		{Name: "Solidus", Data: UnescapeTestCase{"\\/", "/"}},
		{Name: "ControlCharacterInMiddle", Data: UnescapeTestCase{"hello\x19world", "hello"}},
		{Name: "MultipleUnicode", Data: UnescapeTestCase{"\\u0041\\u0042", "AB"}},
		{Name: "HighSurrogateFollowedByNonSurrogate", Data: UnescapeTestCase{"\\uD83D\\u0041", "\uFFFDA"}},
		{Name: "HighSurrogateFollowedByInvalidUnicode", Data: UnescapeTestCase{"\\uD83D\\uXXXX", "\uFFFD\u0000"}},
		{Name: "Unescape", Data: UnescapeTestCase{`{\"name\":null,\"text\":\"您好，我该怎么称呼您呢？\"}`, `{"name":null,"text":"您好，我该怎么称呼您呢？"}`}},
	}

    tt.RunTests(tests, func(subTt *zlsgo.TestUtil, tc zlsgo.TestCase) {
        d := tc.Data.(UnescapeTestCase)
        result := unescape(d.input)
        subTt.Equal(d.expected, result)
    })
}

func TestParseString(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tests := []struct {
		input    string
		expected string
		escaped  bool
	}{
		{"\"simple\"", "simple", false},
		{"\"escaped\\\"quote\"", "escaped\"quote", true},
		{"\"unicode\\u0041\"", "unicodeA", true},
		{"\"invalid\\x\"", "invalid", true},
		{"\"mixed\\t\\u0041\"", "mixed\tA", true},
		{"\"surrogate\\uD83D\\uDE00\"", "surrogate😀", true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			i, val, vesc, ok := parseString(test.input, 1)
			tt.Equal(true, ok)
			tt.Equal(test.escaped, vesc)
			// parseString returns the full string including quotes
			tt.Equal(test.input, val)
			tt.Equal(len(test.input), i)

			if vesc {
				unescaped := unescape(val[1 : len(val)-1])
				tt.Equal(test.expected, unescaped)
			} else {
				// For non-escaped strings, the content should match expected
				tt.Equal(test.expected, val[1:len(val)-1])
			}
		})
	}

	// Test invalid strings
	invalidTests := []string{
		"\"unterminated",
		"\"invalid\\",
		"\"invalid\\u12",
	}

	for _, test := range invalidTests {
		t.Run(test, func(t *testing.T) {
			_, _, _, ok := parseString(test, 1)
			tt.Equal(false, ok)
		})
	}
}

func TestNameOfLastAndSimpleName(t *testing.T) {
	cases := []struct {
		path string
		exp  string
	}{
		{"foo.bar", "bar"},
		{"alpha|beta", "beta"},
		{"foo\\.bar", "foo\\.bar"},
	}
	for _, tc := range cases {
		if got := nameOfLast(tc.path); got != tc.exp {
			t.Fatalf("nameOfLast(%q)=%q want %q", tc.path, got, tc.exp)
		}
	}

	if !isSimpleName("simpleName") {
		t.Fatal("expected simple name to be allowed")
	}
	if isSimpleName("pipe|name") {
		t.Fatal("expected pipe to mark name complex")
	}
	if isSimpleName("bad\nname") {
		t.Fatal("expected control characters to mark name complex")
	}
}

func TestAppendJSONString(t *testing.T) {
	if got := string(appendJSONString(nil, "plain")); got != "\"plain\"" {
		t.Fatalf("appendJSONString plain got %s", got)
	}
	encoded := appendJSONString([]byte{'p'}, "emoji😀")
	if string(encoded[1:]) == "emoji😀" {
		t.Fatal("expected encoded string to escape non-ascii content")
	}
}

func TestValidNullHelper(t *testing.T) {
	if _, ok := validnull([]byte("ull"), 0); !ok {
		t.Fatal("expected validnull to accept ull sequence")
	}
	if _, ok := validnull([]byte("ulx"), 0); ok {
		t.Fatal("expected validnull to reject invalid sequence")
	}
}

func TestArrayQueriesAndPipes(t *testing.T) {
	data := `{"users":[{"name":"Ann","age":25},{"name":"Bob","age":20},{"name":"Eve","age":30}]}`
	all := Get(data, "users.#(age>=25)#.name")
	if all.String() != `["Ann","Eve"]` {
		t.Fatalf("unexpected all query result %s", all.String())
	}
	first := Get(data, "users.#(age>=25).name")
	if first.String() != "Ann" {
		t.Fatalf("unexpected first query result %s", first.String())
	}
	count := Get(data, "users.#")
	if count.Int() != 3 {
		t.Fatalf("unexpected count %d", count.Int())
	}
	SetModifiersState(true)
	defer SetModifiersState(false)
	reversed := Get(data, "users.#.age|@reverse")
	if reversed.String() != `[30,20,25]` {
		t.Fatalf("unexpected reversed result %s", reversed.String())
	}
}

func TestSplitPossiblePipe(t *testing.T) {
	left, right, ok := splitPossiblePipe(".age|@reverse")
	if !ok || left != ".age" || right != "@reverse" {
		t.Fatalf("splitPossiblePipe failed: %v %q %q", ok, left, right)
	}
	if _, _, ok = splitPossiblePipe(`.value\|literal`); ok {
		t.Fatal("expected escaped pipe not to split")
	}
}

func TestParseSubSelectors(t *testing.T) {
	sels, rest, ok := parseSubSelectors(`{"label":.users.0.name,.users.list[0]}`)
	if !ok || rest != "" || len(sels) != 2 {
		t.Fatalf("unexpected selector parse result ok=%v rest=%q len=%d", ok, rest, len(sels))
	}
	if sels[0].name != `"label"` || sels[0].path != ".users.0.name" {
		t.Fatalf("unexpected first selector %#v", sels[0])
	}
	if sels[1].name != "" || sels[1].path != ".users.list[0]" {
		t.Fatalf("unexpected second selector %#v", sels[1])
	}
	_, trailing, ok := parseSubSelectors(`{.users.0.name}.extra`)
	if !ok || trailing != ".extra" {
		t.Fatalf("expected trailing path, got ok=%v tail=%q", ok, trailing)
	}
}

func TestQueryMatchesVariants(t *testing.T) {
	data := `{"items":[{"name":"Anne","active":true,"score":5,"meta":{"flag":true}},{"name":"Bob","active":false,"score":10,"meta":{"flag":false}},{"name":"Carla","active":true,"score":8,"meta":{"flag":true}}]}`
	meta := Get(data, "items.#(meta.flag==true)#.name")
	if meta.String() != `["Anne","Carla"]` {
		t.Fatalf("meta flag filter unexpected %s", meta.String())
	}
	contains := Get(data, `items.#(name!%"B*")#.name`)
	if contains.String() != `["Anne","Carla"]` {
		t.Fatalf("contains filter unexpected %s", contains.String())
	}
	gt := Get(data, "items.#(active> false)#.name")
	if gt.String() != `["Anne","Carla"]` {
		t.Fatalf("bool greater filter unexpected %s", gt.String())
	}
	lt := Get(data, "items.#(active< true)#.name")
	if lt.String() != `["Bob"]` {
		t.Fatalf("bool less filter unexpected %s", lt.String())
	}
	smaller := Get(data, "items.#(score<8)#.score")
	if smaller.String() != `[5]` {
		t.Fatalf("numeric less filter unexpected %s", smaller.String())
	}
}

func TestParseArrayPathDirect(t *testing.T) {
	cases := []struct {
		path  string
		check func(*testing.T, arrayPathResult)
	}{
		{
			path: "3",
			check: func(t *testing.T, res arrayPathResult) {
				if res.part != "3" || res.more || res.piped {
					t.Fatalf("unexpected simple result %#v", res)
				}
			},
		},
		{
			path: "2.rest",
			check: func(t *testing.T, res arrayPathResult) {
				if !res.more || res.path != "rest" {
					t.Fatalf("expected nested path %#v", res)
				}
			},
		},
		{
			path: "1|@reverse",
			check: func(t *testing.T, res arrayPathResult) {
				if !res.piped || res.pipe != "@reverse" {
					t.Fatalf("expected pipe split %#v", res)
				}
			},
		},
		{
			path: "#.name",
			check: func(t *testing.T, res arrayPathResult) {
				if !res.arrch || !res.alogok || res.alogkey != "name" || res.part != "#" {
					t.Fatalf("expected aggregator metadata %#v", res)
				}
			},
		},
		{
			path: "#(count>=2)#.value|@ugly",
			check: func(t *testing.T, res arrayPathResult) {
				if !res.query.on || !res.query.all || res.query.path != "count" || res.query.op != ">=" || res.query.value != "2" {
					t.Fatalf("expected query fields %#v", res)
				}
				if !res.more || res.path != "value|@ugly" {
					t.Fatalf("expected more path with pipe %#v", res)
				}
				if left, right, ok := splitPossiblePipe(res.path); !ok || left != "value" || right != "@ugly" {
					t.Fatalf("expected splitPossiblePipe to detect pipe, got %v %q %q", ok, left, right)
				}
			},
		},
	}
	for _, tc := range cases {
		res := parseArrayPath(tc.path)
		tc.check(t, res)
	}
}

func TestResValueVariants(t *testing.T) {
	json := `{"num":1.5,"bool":true,"obj":{"name":"Anne"},"arr":[1,2]}`
	res := Parse(json)
	if v := res.Get("num").Value(); v != 1.5 {
		t.Fatalf("expected numeric value 1.5 got %v", v)
	}
	if v := res.Get("bool").Value(); v != true {
		t.Fatalf("expected boolean true got %v", v)
	}
	objVal := res.Get("obj").Value()
	objMap, ok := objVal.(map[string]interface{})
	if !ok || objMap["name"] != "Anne" {
		t.Fatalf("expected map with name Anne, got %#v", objVal)
	}
	arrVal := res.Get("arr").Value()
	arrSlice, ok := arrVal.([]interface{})
	if !ok || len(arrSlice) != 2 || arrSlice[0].(float64) != 1 {
		t.Fatalf("unexpected array value %#v", arrVal)
	}
	if res.Get("missing").Value() != nil {
		t.Fatal("expected nil for missing value")
	}
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

type QueryTestCase struct {
	json   string
	path   string
	expect string
}

func TestQueryMatchesEdgeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

    tests := []zlsgo.TestCase{
		{
			Name: "string less than",
			Data: QueryTestCase{
				`{"items":[{"name":"apple"},{"name":"banana"},{"name":"cherry"}]}`,
				"items.#(name<\"banana\")#.name",
				`["apple"]`,
			},
		},
		{
			Name: "string less than or equal",
			Data: QueryTestCase{
				`{"items":[{"name":"apple"},{"name":"banana"},{"name":"cherry"}]}`,
				"items.#(name<=\"banana\")#.name",
				`["apple","banana"]`,
			},
		},
		{
			Name: "string greater than",
			Data: QueryTestCase{
				`{"items":[{"name":"apple"},{"name":"banana"},{"name":"cherry"}]}`,
				"items.#(name>\"banana\")#.name",
				`["cherry"]`,
			},
		},
		{
			Name: "string greater than or equal",
			Data: QueryTestCase{
				`{"items":[{"name":"apple"},{"name":"banana"},{"name":"cherry"}]}`,
				"items.#(name>=\"banana\")#.name",
				`["banana","cherry"]`,
			},
		},
		{
			Name: "number not equal",
			Data: QueryTestCase{
				`{"items":[{"val":1},{"val":2},{"val":3}]}`,
				"items.#(val!=2)#.val",
				`[1,3]`,
			},
		},
		{
			Name: "number less than or equal",
			Data: QueryTestCase{
				`{"items":[{"val":1},{"val":2},{"val":3}]}`,
				"items.#(val<=2)#.val",
				`[1,2]`,
			},
		},
		{
			Name: "bool equals false",
			Data: QueryTestCase{
				`{"items":[{"active":true},{"active":false}]}`,
				"items.#(active==false)#.active",
				`[false]`,
			},
		},
		{
			Name: "bool not equals false",
			Data: QueryTestCase{
				`{"items":[{"active":true},{"active":false}]}`,
				"items.#(active!=false)#.active",
				`[true]`,
			},
		},
		{
			Name: "bool greater than or equal true",
			Data: QueryTestCase{
				`{"items":[{"active":true},{"active":false}]}`,
				"items.#(active>=true)#.active",
				`[true]`,
			},
		},
		{
			Name: "bool less than or equal true",
			Data: QueryTestCase{
				`{"items":[{"active":true},{"active":false}]}`,
				"items.#(active<=false)#.active",
				`[false]`,
			},
		},
		{
			Name: "pattern match",
			Data: QueryTestCase{
				`{"items":[{"name":"test123"},{"name":"foo"},{"name":"test456"}]}`,
				`items.#(name%"test*")#.name`,
				`["test123","test456"]`,
			},
		},
		{
			Name: "pattern not match",
			Data: QueryTestCase{
				`{"items":[{"name":"test123"},{"name":"foo"},{"name":"test456"}]}`,
				`items.#(name!%"test*")#.name`,
				`["foo"]`,
			},
		},
	}

    tt.RunTests(tests, func(subTt *zlsgo.TestUtil, tc zlsgo.TestCase) {
        d := tc.Data.(QueryTestCase)
        result := Get(d.json, d.path)
        subTt.Equal(d.expect, result.String())
    })
}

func TestSplitPossiblePipeEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		hasLeft  string
		hasRight string
		hasSplit bool
	}{
		{
			name:     "no pipe",
			path:     "users.0.name",
			hasSplit: false,
		},
		{
			name:     "simple pipe",
			path:     "users|@reverse",
			hasLeft:  "users",
			hasRight: "@reverse",
			hasSplit: true,
		},
		{
			name:     "escaped pipe in path",
			path:     `users\.name\|escaped`,
			hasSplit: false,
		},
		{
			name:     "pipe after query with brackets",
			path:     "users.#[name==\"test\"]#.id|@ugly",
			hasLeft:  "users.#[name==\"test\"]#.id",
			hasRight: "@ugly",
			hasSplit: true,
		},
		{
			name:     "pipe after query with parens",
			path:     "users.#(name==\"test\")#.id|@pretty",
			hasLeft:  "users.#(name==\"test\")#.id",
			hasRight: "@pretty",
			hasSplit: true,
		},
		{
			name:     "nested brackets with pipe",
			path:     `users.#[data.items[0]=="test"]#.id|@reverse`,
			hasLeft:  `users.#[data.items[0]=="test"]#.id`,
			hasRight: "@reverse",
			hasSplit: true,
		},
		{
			name:     "nested parens with pipe",
			path:     `users.#(data.check(val))#.id|@reverse`,
			hasLeft:  `users.#(data.check(val))#.id`,
			hasRight: "@reverse",
			hasSplit: true,
		},
		{
			name:     "nested string with pipe inside query",
			path:     `users.#(name=="test|pipe")#.id|@ugly`,
			hasLeft:  `users.#(name=="test|pipe")#.id`,
			hasRight: "@ugly",
			hasSplit: true,
		},
		{
			name:     "path ends with dot-hash",
			path:     "users.#",
			hasSplit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			left, right, ok := splitPossiblePipe(tt.path)
			if ok != tt.hasSplit {
				t.Errorf("splitPossiblePipe(%q) ok = %v, want %v", tt.path, ok, tt.hasSplit)
			}
			if ok {
				if left != tt.hasLeft {
					t.Errorf("splitPossiblePipe(%q) left = %q, want %q", tt.path, left, tt.hasLeft)
				}
				if right != tt.hasRight {
					t.Errorf("splitPossiblePipe(%q) right = %q, want %q", tt.path, right, tt.hasRight)
				}
			}
		})
	}
}

func TestParseArrayPathComplexQueries(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		check func(t *testing.T, res arrayPathResult)
	}{
		{
			name: "query with brackets",
			path: "#[age>25]",
			check: func(t *testing.T, res arrayPathResult) {
				if !res.query.on || res.query.path != "age" || res.query.op != ">" || res.query.value != "25" {
					t.Errorf("unexpected query result: %+v", res)
				}
			},
		},
		{
			name: "query with parentheses all",
			path: "#(status==\"active\")#",
			check: func(t *testing.T, res arrayPathResult) {
				if !res.query.on || !res.query.all {
					t.Errorf("expected query all: %+v", res)
				}
			},
		},
		{
			name: "query with nested path",
			path: "#(user.name==\"test\").data",
			check: func(t *testing.T, res arrayPathResult) {
				if !res.query.on || res.query.path != "user.name" {
					t.Errorf("expected nested query path: %+v", res)
				}
				if !res.more || res.path != "data" {
					t.Errorf("expected more path: %+v", res)
				}
			},
		},
		{
			name: "array alog key",
			path: "#.items",
			check: func(t *testing.T, res arrayPathResult) {
				if !res.alogok || res.alogkey != "items" {
					t.Errorf("expected alog key items: %+v", res)
				}
			},
		},
		{
			name: "array index with more path",
			path: "5.nested.value",
			check: func(t *testing.T, res arrayPathResult) {
				if res.part != "5" || !res.more || res.path != "nested.value" {
					t.Errorf("unexpected indexed path: %+v", res)
				}
			},
		},
		{
			name: "piped result",
			path: "3|@reverse",
			check: func(t *testing.T, res arrayPathResult) {
				if !res.piped || res.pipe != "@reverse" {
					t.Errorf("expected pipe: %+v", res)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := parseArrayPath(tt.path)
			tt.check(t, res)
		})
	}
}

func TestIntUintEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		path     string
		wantInt  int
		wantUint uint
	}{
		{
			name:     "negative int",
			json:     `{"val":-123}`,
			path:     "val",
			wantInt:  -123,
			wantUint: 0,
		},
		{
			name:     "large positive",
			json:     `{"val":999999}`,
			path:     "val",
			wantInt:  999999,
			wantUint: 999999,
		},
		{
			name:     "zero",
			json:     `{"val":0}`,
			path:     "val",
			wantInt:  0,
			wantUint: 0,
		},
		{
			name:     "string number",
			json:     `{"val":"42"}`,
			path:     "val",
			wantInt:  42,
			wantUint: 42,
		},
		{
			name:     "invalid string",
			json:     `{"val":"abc"}`,
			path:     "val",
			wantInt:  0,
			wantUint: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Get(tt.json, tt.path)
			if res.Int() != tt.wantInt {
				t.Errorf("Int() = %d, want %d", res.Int(), tt.wantInt)
			}
			if res.Uint() != tt.wantUint {
				t.Errorf("Uint() = %d, want %d", res.Uint(), tt.wantUint)
			}
		})
	}
}

func TestSliceEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		json  string
		path  string
		check func(t *testing.T, res *Res)
	}{
		{
			name: "array of mixed types",
			json: `{"arr":[1,"two",true,null,{"key":"val"}]}`,
			path: "arr",
			check: func(t *testing.T, res *Res) {
				slice := res.Slice()
				if len(slice) != 5 {
					t.Errorf("expected 5 items, got %d", len(slice))
				}
			},
		},
		{
			name: "non-array",
			json: `{"val":"notarray"}`,
			path: "val",
			check: func(t *testing.T, res *Res) {
				slice := res.Slice()
				if len(slice) != 0 {
					t.Errorf("expected empty slice for non-array, got %d items", len(slice))
				}
			},
		},
		{
			name: "empty array",
			json: `{"arr":[]}`,
			path: "arr",
			check: func(t *testing.T, res *Res) {
				slice := res.Slice()
				if len(slice) != 0 {
					t.Errorf("expected empty slice, got %d items", len(slice))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Get(tt.json, tt.path)
			tt.check(t, res)
		})
	}
}

func TestArrayEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		json  string
		path  string
		check func(t *testing.T, res *Res)
	}{
		{
			name: "nested array",
			json: `{"data":[[1,2],[3,4]]}`,
			path: "data",
			check: func(t *testing.T, res *Res) {
				arr := res.Array()
				if len(arr) != 2 {
					t.Errorf("expected 2 items, got %d", len(arr))
				}
			},
		},
		{
			name: "array in non-array value",
			json: `{"val":123}`,
			path: "val",
			check: func(t *testing.T, res *Res) {
				arr := res.Array()
				// Non-array values return a single-element array
				if len(arr) != 1 {
					t.Errorf("expected single-element array for non-array, got %d", len(arr))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Get(tt.json, tt.path)
			tt.check(t, res)
		})
	}
}

func TestValueEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		json  string
		path  string
		check func(t *testing.T, val interface{})
	}{
		{
			name: "null value",
			json: `{"val":null}`,
			path: "val",
			check: func(t *testing.T, val interface{}) {
				if val != nil {
					t.Errorf("expected nil for null, got %v", val)
				}
			},
		},
		{
			name: "string value",
			json: `{"val":"hello"}`,
			path: "val",
			check: func(t *testing.T, val interface{}) {
				if val != "hello" {
					t.Errorf("expected 'hello', got %v", val)
				}
			},
		},
		{
			name: "nested object",
			json: `{"val":{"nested":{"deep":"value"}}}`,
			path: "val",
			check: func(t *testing.T, val interface{}) {
				m, ok := val.(map[string]interface{})
				if !ok {
					t.Errorf("expected map, got %T", val)
					return
				}
				nested, ok := m["nested"].(map[string]interface{})
				if !ok || nested["deep"] != "value" {
					t.Errorf("unexpected nested structure: %+v", m)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Get(tt.json, tt.path)
			tt.check(t, res.Value())
		})
	}
}

func TestParseEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		json  string
		check func(t *testing.T, res *Res)
	}{
		{
			name: "incomplete number at end",
			json: `{"val":12`,
			check: func(t *testing.T, res *Res) {
				val := res.Get("val").Int()
				if val != 12 {
					t.Errorf("expected to parse incomplete number, got %d", val)
				}
			},
		},
		{
			name: "incomplete string",
			json: `{"val":"test`,
			check: func(t *testing.T, res *Res) {
				val := res.Get("val").String()
				// Incomplete strings return empty
				if val != "" {
					t.Logf("incomplete string parsed as: %q", val)
				}
			},
		},
		{
			name: "multiple values",
			json: `{"a":1}{"b":2}`,
			check: func(t *testing.T, res *Res) {
				if res.Get("a").Int() != 1 {
					t.Errorf("expected first object to parse")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Parse(tt.json)
			tt.check(t, res)
		})
	}
}

func TestForEachEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		json  string
		path  string
		check func(t *testing.T)
	}{
		{
			name: "stop iteration",
			json: `{"items":[1,2,3,4,5]}`,
			path: "items",
			check: func(t *testing.T) {
				count := 0
				Get(`{"items":[1,2,3,4,5]}`, "items").ForEach(func(key, value *Res) bool {
					count++
					return count < 3 // Stop after 3 items
				})
				if count != 3 {
					t.Errorf("expected to stop at 3, got %d", count)
				}
			},
		},
		{
			name: "empty object",
			json: `{"obj":{}}`,
			path: "obj",
			check: func(t *testing.T) {
				count := 0
				Get(`{"obj":{}}`, "obj").ForEach(func(key, value *Res) bool {
					count++
					return true
				})
				if count != 0 {
					t.Errorf("expected 0 iterations for empty object, got %d", count)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t)
		})
	}
}

func TestGetMultipleEdgeCases(t *testing.T) {
	json := `{"a":1,"b":2,"c":3}`
	results := GetMultiple(json, "a", "b", "c")

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	if results[0].Int() != 1 {
		t.Errorf("expected a=1, got %d", results[0].Int())
	}
	if results[1].Int() != 2 {
		t.Errorf("expected b=2, got %d", results[1].Int())
	}
	if results[2].Int() != 3 {
		t.Errorf("expected c=3, got %d", results[2].Int())
	}
}

func TestForEachLineEdgeCases(t *testing.T) {
	json := `{"a":1}
{"b":2}
{"c":3}`

	count := 0
	ForEachLine(json, func(line *Res) bool {
		count++
		return true
	})

	if count != 3 {
		t.Errorf("expected 3 lines, got %d", count)
	}

	// Test early termination
	count = 0
	ForEachLine(json, func(line *Res) bool {
		count++
		return count < 2
	})

	if count != 2 {
		t.Errorf("expected early termination at 2, got %d", count)
	}
}

func TestParseComplexPaths(t *testing.T) {
	json := `{
		"users": [
			{"name": "Alice", "age": 30, "active": true},
			{"name": "Bob", "age": 25, "active": false},
			{"name": "Charlie", "age": 35, "active": true}
		],
		"metadata": {
			"total": 3,
			"page": 1
		}
	}`

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "array length",
			path:     "users.#",
			expected: "3",
		},
		{
			name:     "nested object",
			path:     "metadata.total",
			expected: "3",
		},
		{
			name:     "array index",
			path:     "users.0.name",
			expected: "Alice",
		},
		{
			name:     "array last element",
			path:     "users.2.name",
			expected: "Charlie",
		},
		{
			name:     "query equals",
			path:     "users.#(name==\"Bob\").age",
			expected: "25",
		},
		{
			name:     "query all",
			path:     "users.#(active==true)#.name",
			expected: `["Alice","Charlie"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(json, tt.path)
			if result.String() != tt.expected {
				t.Errorf("Get(%q) = %q, want %q", tt.path, result.String(), tt.expected)
			}
		})
	}
}

func TestToNumEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		path     string
		expected float64
	}{
		{
			name:     "integer",
			json:     `{"val":42}`,
			path:     "val",
			expected: 42,
		},
		{
			name:     "negative integer",
			json:     `{"val":-42}`,
			path:     "val",
			expected: -42,
		},
		{
			name:     "float",
			json:     `{"val":3.14}`,
			path:     "val",
			expected: 3.14,
		},
		{
			name:     "scientific notation",
			json:     `{"val":1.23e10}`,
			path:     "val",
			expected: 1.23e10,
		},
		{
			name:     "zero",
			json:     `{"val":0}`,
			path:     "val",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(tt.json, tt.path).Float64()
			if result != tt.expected {
				t.Errorf("Float64() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseObjectPathEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		path     string
		expected string
	}{
		{
			name:     "simple key",
			json:     `{"key":"value"}`,
			path:     "key",
			expected: "value",
		},
		{
			name:     "nested keys",
			json:     `{"a":{"b":{"c":"deep"}}}`,
			path:     "a.b.c",
			expected: "deep",
		},
		{
			name:     "key with spaces",
			json:     `{"key with spaces":"value"}`,
			path:     "key with spaces",
			expected: "value",
		},
		{
			name:     "unicode key",
			json:     `{"键":"值"}`,
			path:     "键",
			expected: "值",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(tt.json, tt.path)
			if result.String() != tt.expected {
				t.Errorf("Get(%q) = %q, want %q", tt.path, result.String(), tt.expected)
			}
		})
	}
}

func TestIsOptimisticPath(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"simple", true},
		{"simple.path", true},
		{"simple.path.nested", true},
		{"a.b.c.d.e", true},
		{"path0.with1.numbers2", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isOptimisticPath(tt.path)
			if result != tt.expected {
				t.Errorf("isOptimisticPath(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}
