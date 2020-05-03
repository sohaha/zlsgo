package zjson

import (
	"encoding/json"
	"testing"

	"github.com/sohaha/zlsgo/zstring"
)

func BenchmarkUnmarshal(b *testing.B) {
	var demoData Demo
	demoByte := zstring.String2Bytes(demo)
	for i := 0; i < b.N; i++ {
		_ = Unmarshal(demoByte, &demoData)
	}
}

func BenchmarkGolangUnmarshal(b *testing.B) {
	var demoData Demo
	demoByte := zstring.String2Bytes(demo)
	for i := 0; i < b.N; i++ {
		_ = json.Unmarshal(demoByte, &demoData)
	}
}

func BenchmarkMarshal(b *testing.B) {
	var demoData Demo
	_ = Unmarshal(zstring.String2Bytes(demo), &demoData)
	for i := 0; i < b.N; i++ {
		_, _ = Marshal(demoData)
	}
}

func BenchmarkStringify(b *testing.B) {
	var demoData Demo
	_ = Unmarshal(zstring.String2Bytes(demo), &demoData)
	for i := 0; i < b.N; i++ {
		Stringify(demoData)
	}
}

func BenchmarkGolangMarshal(b *testing.B) {
	var demoData Demo
	_ = Unmarshal(zstring.String2Bytes(demo), &demoData)
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(demoData)
	}
}
