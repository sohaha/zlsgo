package zjson

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/sohaha/zlsgo/zstring"
)

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
