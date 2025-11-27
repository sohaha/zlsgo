package zutil

import (
	"strconv"
	"strings"
	"unsafe"
)

func UnescapeHTML(s string) string {
	s = strings.Replace(s, "\\u003c", "<", -1)
	s = strings.Replace(s, "\\u003e", ">", -1)
	return strings.Replace(s, "\\u0026", "&", -1)
}

func KeySignature(key interface{}) string {
	var b strings.Builder
	b.Grow(64)
	appendKeyRepr(&b, key)
	return b.String()
}

func appendKeyRepr(b *strings.Builder, key interface{}) {
	switch v := key.(type) {
	case string:
		b.WriteByte('s')
		b.WriteString(v)
	case int:
		b.WriteByte('i')
		appendInt(b, int64(v))
	case int8:
		b.WriteByte('a')
		appendInt(b, int64(v))
	case int16:
		b.WriteByte('b')
		appendInt(b, int64(v))
	case int32:
		b.WriteByte('c')
		appendInt(b, int64(v))
	case int64:
		b.WriteByte('d')
		appendInt(b, v)
	case uint:
		b.WriteByte('u')
		appendUint(b, uint64(v), 10)
	case uint8:
		b.WriteByte('v')
		appendUint(b, uint64(v), 10)
	case uint16:
		b.WriteByte('w')
		appendUint(b, uint64(v), 10)
	case uint32:
		b.WriteByte('x')
		appendUint(b, uint64(v), 10)
	case uint64:
		b.WriteByte('y')
		appendUint(b, v, 10)
	case uintptr:
		b.WriteByte('p')
		appendUint(b, uint64(v), 16)
	case unsafe.Pointer:
		b.WriteByte('P')
		appendUint(b, uint64(uintptr(v)), 16)
	case float32:
		b.WriteByte('f')
		appendFloat(b, float64(v), 32)
	case float64:
		b.WriteByte('F')
		appendFloat(b, v, 64)
	case complex64:
		b.WriteByte('g')
		appendFloat(b, float64(real(v)), 32)
		b.WriteByte(',')
		appendFloat(b, float64(imag(v)), 32)
	case complex128:
		b.WriteByte('G')
		appendFloat(b, real(v), 64)
		b.WriteByte(',')
		appendFloat(b, imag(v), 64)
	default:
		b.WriteByte('?')
	}
}

func appendInt(b *strings.Builder, v int64) {
	var buf [32]byte
	b.Write(strconv.AppendInt(buf[:0], v, 10))
}

func appendUint(b *strings.Builder, v uint64, base int) {
	var buf [32]byte
	b.Write(strconv.AppendUint(buf[:0], v, base))
}

func appendFloat(b *strings.Builder, v float64, bitSize int) {
	var buf [64]byte
	b.Write(strconv.AppendFloat(buf[:0], v, 'g', -1, bitSize))
}
