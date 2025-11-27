package zutil_test

import (
	"strings"
	"testing"
	"unsafe"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zutil"
)

func TestUnescapeHTML(t *testing.T) {
	tt := zlsgo.NewTest(t)
	s := zutil.UnescapeHTML("")
	tt.Log(s)
}

func TestKeySignature(t *testing.T) {
	tt := zlsgo.NewTest(t)
	var ptrTarget int
	tests := []struct {
		name   string
		input  interface{}
		want   string
		assert func(got string)
	}{
		{name: "string", input: "a", want: "sa"},
		{name: "int", input: 123, want: "i123"},
		{name: "int8", input: int8(-9), want: "a-9"},
		{name: "int16", input: int16(42), want: "b42"},
		{name: "int32", input: int32(-1024), want: "c-1024"},
		{name: "int64", input: int64(999), want: "d999"},
		{name: "uint", input: uint(7), want: "u7"},
		{name: "uint8", input: uint8(8), want: "v8"},
		{name: "uint16", input: uint16(16), want: "w16"},
		{name: "uint32", input: uint32(32), want: "x32"},
		{name: "uint64", input: uint64(64), want: "y64"},
		{name: "uintptr", input: uintptr(0x123abc), want: "p123abc"},
		{name: "float32", input: float32(1.5), want: "f1.5"},
		{name: "float64", input: float64(-2.25), want: "F-2.25"},
		{name: "complex64", input: complex64(1.25 + 2i), want: "g1.25,2"},
		{name: "complex128", input: complex128(-3.5 + 0.5i), want: "G-3.5,0.5"},
		{
			name:  "unsafePointer",
			input: unsafe.Pointer(&ptrTarget),
			assert: func(got string) {
				tt.EqualTrue(strings.HasPrefix(got, "P"))
			},
		},
		{name: "default", input: struct{}{}, want: "?"},
	}

	for _, tc := range tests {
		got := zutil.KeySignature(tc.input)
		if tc.assert != nil {
			tc.assert(got)
			continue
		}
		tt.Equal(tc.want, got)
	}
}
