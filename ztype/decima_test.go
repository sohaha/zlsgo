package ztype_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/ztype"
)

func TestDecimal(tt *testing.T) {
	t := zlsgo.NewTest(tt)

	t.Equal("10100", strconv.FormatInt(20, 2))
	t.Equal("10100", fmt.Sprintf("%b", 20))
	t.Equal("10100", ztype.DecimalToAny(20, 2))

	t.Equal(strconv.FormatInt(20, 2), ztype.DecimalToAny(20, 2))
	t.Equal(strconv.FormatInt(20, 8), ztype.DecimalToAny(20, 8))
	t.Equal(strconv.FormatInt(20, 16), ztype.DecimalToAny(20, 16))
	t.Equal(strconv.FormatInt(20, 32), ztype.DecimalToAny(20, 32))
	t.Equal(strconv.FormatInt(20, 32), ztype.DecimalToAny(20, 31))

	t.Equal("10100", ztype.DecimalToAny(20, 2))
	t.Equal("2Bi", ztype.DecimalToAny(10000, 62))
	t.Equal(10000, ztype.AnyToDecimal("2Bi", 62))
	t.Equal(0, ztype.AnyToDecimal("20", 0))
	t.Equal(2281, ztype.AnyToDecimal("2Bi", 31))

	tt.Log(ztype.DecimalToAny(65, 66))
	tt.Log(strconv.FormatInt(65, 9), ztype.DecimalToAny(65, 9))
	t.Equal("10011100010000", ztype.DecimalToAny(10000, 2))
}

func BenchmarkDecimalStrconv2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v := strconv.FormatInt(20, 2)
		_ = v
	}
}

func BenchmarkDecimalZtype2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v := ztype.DecimalToAny(20, 2)
		_ = v
	}
}

func BenchmarkDecimalStrconv8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v := strconv.FormatInt(20, 8)
		_, _ = strconv.ParseInt(v, 8, 64)
	}
}

func BenchmarkDecimalZtype8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v := ztype.DecimalToAny(20, 8)
		_ = ztype.AnyToDecimal(v, 8)
	}
}

func BenchmarkDecimalStrconv16(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v := strconv.FormatInt(20, 16)
		_, _ = strconv.ParseInt(v, 16, 64)
	}
}

func BenchmarkDecimalZtype16(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v := ztype.DecimalToAny(20, 16)
		_ = ztype.AnyToDecimal(v, 16)
	}
}

func BenchmarkDecimalStrconv32(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v := strconv.FormatInt(20, 32)
		_, _ = strconv.ParseInt(v, 32, 64)
	}
}

func BenchmarkDecimalZtype32(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v := ztype.DecimalToAny(20, 32)
		_ = ztype.AnyToDecimal(v, 32)
	}
}
