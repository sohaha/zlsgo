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

func BenchmarkUnmarshal(b *testing.B) {
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

func BenchmarkMarshal(b *testing.B) {
	var demoData Demo
	_ = Unmarshal(zstring.String2Bytes(demo), &demoData)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Marshal(demoData)
	}
}

func BenchmarkSetMarshal(b *testing.B) {
	var demoData Demo
	_ = Unmarshal(zstring.String2Bytes(demo), &demoData)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SetBytesOptions([]byte("{}"), "", demoData, &Options{
			Optimistic: true,
		})
	}
}

func BenchmarkSet2Marshal(b *testing.B) {
	var demoData Demo
	_ = Unmarshal(zstring.String2Bytes(demo), &demoData)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Set("{}", "", demoData)
	}
}

func BenchmarkGolangMarshal(b *testing.B) {
	var demoData Demo
	_ = Unmarshal(zstring.String2Bytes(demo), &demoData)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(demoData)
	}
}

func BenchmarkStringify(b *testing.B) {
	var demoData Demo
	_ = Unmarshal(zstring.String2Bytes(demo), &demoData)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Stringify(demoData)
	}
}
