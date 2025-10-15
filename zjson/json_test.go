package zjson

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zstring"
)

func TestM(t *testing.T) {
	tt := zlsgo.NewTest(t)

	b := []byte(`{
    "id": "chatcmpl-C1BWndEIRpjhAaF3YOL5fvUtKXW4F",
    "object": "chat.completion",
    "created": 1754398677,
    "model": "gpt-4.1-mini-2025-04-14",
    "choices": [
      {
        "index": 0,
        "message": {
          "role": "assistant",
          "content": "{\"name\":null,\"gender\":null,\"age\":null,\"phone\":null,\"done\":false,\"text\":\"您好！我是本次的登记专员，想和您核对几个基本信息，方便吗？首先，我该怎么称呼您呢？\"}",
          "refusal": null,
          "annotations": []
        },
        "logprobs": null,
        "finish_reason": "stop"
      }
    ],
    "usage": {
      "prompt_tokens": 1999,
      "completion_tokens": 55,
      "total_tokens": 2054,
      "prompt_tokens_details": {
        "cached_tokens": 1664,
        "audio_tokens": 0
      },
      "completion_tokens_details": {
        "reasoning_tokens": 0,
        "audio_tokens": 0,
        "accepted_prediction_tokens": 0,
        "rejected_prediction_tokens": 0
      }
    },
    "service_tier": "default",
    "system_fingerprint": "fp_658b958c37"
  }`)
	j := Parse(string(b))
	tt.Log("33")
	j.Set("xxx", "是我呀")
	tt.Log(j.Get("choices.0.message").String())
	tt.Log(2, j.Get("choices.0.message").Get("content").String())
	tt.Log(1111, j.Get("choices.0.message.content").String())
	tt.Log(j.Get("xxx").String())
}

func TestMatch(t *testing.T) {
	tt := zlsgo.NewTest(t)
	j := Parse(demo)
	nj := j.MatchKeys([]string{"time", "friends"})
	tt.Log(j)
	tt.Log(nj)

	j = Parse("")
	nj = j.MatchKeys([]string{"time", "friends"})
	tt.Log(j)
	tt.Log(nj)
}

func getBigJSON() string {
	s := ""
	for i := 0; i < 10000; i++ {
		s, _ = Set(s, strconv.Itoa(i), zstring.Rand(10))
	}
	return s
}

type ValidTestCase struct {
	json  string
	valid bool
}

func TestValidEdgeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tests := []zlsgo.TestCase[ValidTestCase]{
		{Name: "ValidObject", Data: ValidTestCase{`{"a":"b"}`, true}},
		{Name: "MissingColon", Data: ValidTestCase{`{"a" 1}`, false}},
		{Name: "TrailingComma", Data: ValidTestCase{`{"a":1,}`, false}},
		{Name: "BadArrayComma", Data: ValidTestCase{`[1,,2]`, false}},
		{Name: "BadArraySpace", Data: ValidTestCase{`[1 2]`, false}},
		{Name: "BadStringEscape", Data: ValidTestCase{`{"a":"\x"}`, false}},
		{Name: "ControlInString", Data: ValidTestCase{"{\"a\":\"line\nbreak\"}", false}},
		{Name: "ValidUnicodeEscape", Data: ValidTestCase{`{"a":"\u0041"}`, true}},
		{Name: "InvalidUnicodeDigits", Data: ValidTestCase{`{"a":"\u00ZG"}`, false}},
		{Name: "BadNumberExponent", Data: ValidTestCase{`{"a":1e+}`, false}},
	}

	zlsgo.RunTests(tt, tests, func(subTt *zlsgo.TestUtil, tc zlsgo.TestCase[ValidTestCase]) {
		got := Valid(tc.Data.json)
		subTt.Equal(tc.Data.valid, got)
	})
}

func TestJSONSyntaxErrorString(t *testing.T) {
	tt := zlsgo.NewTest(t)
	err := (&JSONSyntaxError{Message: "unexpected token", Position: 12}).Error()
	tt.EqualTrue(err != "" && err[len(err)-1] != ':')
}

func BenchmarkUnmarshal1(b *testing.B) {
	var demoData Demo
	demoByte := zstring.String2Bytes(demo)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Unmarshal(demoByte, &demoData)
	}
}

func BenchmarkGolangUnmarshal(b *testing.B) {
	var demoData Demo
	demoByte := zstring.String2Bytes(demo)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = json.Unmarshal(demoByte, &demoData)
	}
}

type StringTestCase struct {
	input string
	valid bool
}

func TestValidString(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tests := []zlsgo.TestCase[StringTestCase]{
		{Name: "simple string", Data: StringTestCase{`"hello"`, true}},
		{Name: "empty string", Data: StringTestCase{`""`, true}},
		{Name: "escaped quote", Data: StringTestCase{`"hello \"world\""`, true}},
		{Name: "escaped backslash", Data: StringTestCase{`"hello\\world"`, true}},
		{Name: "escaped forward slash", Data: StringTestCase{`"hello\/world"`, true}},
		{Name: "escaped control chars", Data: StringTestCase{`"hello\b\f\n\r\tworld"`, true}},
		{Name: "unicode escape", Data: StringTestCase{`"hello\u0041world"`, true}},
		{Name: "invalid escape", Data: StringTestCase{`"hello\xworld"`, false}},
		{Name: "invalid unicode escape", Data: StringTestCase{`"hello\u00"`, false}},
		{Name: "control character", Data: StringTestCase{"\"hello\nworld\"", false}},
		{Name: "incomplete unicode", Data: StringTestCase{`"hello\u12"`, false}},
		{Name: "invalid unicode char", Data: StringTestCase{`"hello\uGGGG"`, false}},
	}

	zlsgo.RunTests(tt, tests, func(subTt *zlsgo.TestUtil, tc zlsgo.TestCase[StringTestCase]) {
		data := []byte(tc.Data.input)
		_, ok := validstring(data, 1)
		subTt.Equal(tc.Data.valid, ok)
	})
}

type NumberTestCase struct {
	input string
	valid bool
}

func TestValidNumber(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tests := []zlsgo.TestCase[NumberTestCase]{
		{Name: "zero", Data: NumberTestCase{"0", true}},
		{Name: "positive integer", Data: NumberTestCase{"123", true}},
		{Name: "negative integer", Data: NumberTestCase{"-456", true}},
		{Name: "decimal", Data: NumberTestCase{"123.456", true}},
		{Name: "negative decimal", Data: NumberTestCase{"-123.456", true}},
		{Name: "scientific notation", Data: NumberTestCase{"1.23e10", true}},
		{Name: "scientific with plus", Data: NumberTestCase{"1.23e+10", true}},
		{Name: "scientific with minus", Data: NumberTestCase{"1.23e-10", true}},
		{Name: "negative scientific", Data: NumberTestCase{"-1.23e-10", true}},
		{Name: "just zero", Data: NumberTestCase{"0.0", true}},
		{Name: "incomplete decimal", Data: NumberTestCase{"123.", false}},
		{Name: "incomplete exponent", Data: NumberTestCase{"123e", false}},
	}

	zlsgo.RunTests(tt, tests, func(subTt *zlsgo.TestUtil, tc zlsgo.TestCase[NumberTestCase]) {
		data := []byte(tc.Data.input)
		_, ok := validnumber(data, len(data))
		subTt.Equal(tc.Data.valid, ok)
	})
}

type TrueTestCase struct {
	input    string
	startPos int
	valid    bool
}

func TestValidTrue(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tests := []zlsgo.TestCase[TrueTestCase]{
		{Name: "true", Data: TrueTestCase{"true", 1, true}},
		{Name: "tru", Data: TrueTestCase{"tru", 1, false}},
		{Name: "TRUE", Data: TrueTestCase{"TRUE", 1, false}},
	}

	zlsgo.RunTests(tt, tests, func(subTt *zlsgo.TestUtil, tc zlsgo.TestCase[TrueTestCase]) {
		data := []byte(tc.Data.input)
		// validtrue expects index at the position after 't'
		_, ok := validtrue(data, tc.Data.startPos)
		subTt.Equal(tc.Data.valid, ok)
	})
}

type FalseTestCase struct {
	input    string
	startPos int
	valid    bool
}

func TestValidFalse(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tests := []zlsgo.TestCase[FalseTestCase]{
		{Name: "false", Data: FalseTestCase{"false", 1, true}},
		{Name: "fals", Data: FalseTestCase{"fals", 1, false}},
		{Name: "FALSE", Data: FalseTestCase{"FALSE", 1, false}},
	}

	zlsgo.RunTests(tt, tests, func(subTt *zlsgo.TestUtil, tc zlsgo.TestCase[FalseTestCase]) {
		data := []byte(tc.Data.input)
		_, ok := validfalse(data, tc.Data.startPos)
		subTt.Equal(tc.Data.valid, ok)
	})
}

type ParseIntTestCase struct {
	input    string
	expected int
	valid    bool
}

func TestParseInt(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tests := []zlsgo.TestCase[ParseIntTestCase]{
		{Name: "0", Data: ParseIntTestCase{"0", 0, true}},
		{Name: "123", Data: ParseIntTestCase{"123", 123, true}},
		{Name: "-456", Data: ParseIntTestCase{"-456", -456, true}},
		{Name: "999999", Data: ParseIntTestCase{"999999", 999999, true}},
		{Name: "-1", Data: ParseIntTestCase{"-1", -1, true}},
		{Name: "abc", Data: ParseIntTestCase{"abc", 0, false}},
		{Name: "12a", Data: ParseIntTestCase{"12a", 0, false}},
		{Name: "-", Data: ParseIntTestCase{"-", 0, false}},
		{Name: "", Data: ParseIntTestCase{"", 0, false}},
		{Name: "12.34", Data: ParseIntTestCase{"12.34", 0, false}},
	}

	zlsgo.RunTests(tt, tests, func(subTt *zlsgo.TestUtil, tc zlsgo.TestCase[ParseIntTestCase]) {
		n, ok := parseInt(tc.Data.input)
		subTt.Equal(tc.Data.valid, ok)
		if ok {
			subTt.Equal(tc.Data.expected, n)
		}
	})
}

func TestValidPayload(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tests := []zlsgo.TestCase[ValidTestCase]{
		{Name: "valid object", Data: ValidTestCase{`{"a":1}`, true}},
		{Name: "valid array", Data: ValidTestCase{`[1,2,3]`, true}},
		{Name: "valid string", Data: ValidTestCase{`"hello"`, true}},
		{Name: "valid number", Data: ValidTestCase{`123`, true}},
		{Name: "valid bool", Data: ValidTestCase{`true`, true}},
		{Name: "valid null", Data: ValidTestCase{`null`, true}},
		{Name: "invalid unclosed object", Data: ValidTestCase{`{"a":1`, false}},
		{Name: "invalid unclosed array", Data: ValidTestCase{`[1,2`, false}},
		{Name: "invalid trailing comma", Data: ValidTestCase{`{"a":1,}`, false}},
		{Name: "empty string", Data: ValidTestCase{``, false}},
	}

	zlsgo.RunTests(tt, tests, func(subTt *zlsgo.TestUtil, tc zlsgo.TestCase[ValidTestCase]) {
		result := Valid(tc.Data.json)
		subTt.Equal(tc.Data.valid, result)
	})
}
