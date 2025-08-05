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
	j := Parse(demo)
	nj := j.MatchKeys([]string{"time", "friends"})
	t.Log(j)
	t.Log(nj)

	j = Parse("")
	nj = j.MatchKeys([]string{"time", "friends"})
	t.Log(j)
	t.Log(nj)
}

func getBigJSON() string {
	s := ""
	for i := 0; i < 10000; i++ {
		s, _ = Set(s, strconv.Itoa(i), zstring.Rand(10))
	}
	return s
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
