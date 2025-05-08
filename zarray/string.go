//go:build go1.18
// +build go1.18

package zarray

import (
	"strings"

	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztype"
)

// Slice converts a string to a slice.
// If value is not empty, the string will be split into value parts.
func Slice[T comparable](s, sep string, n ...int) []T {
	if s == "" {
		return []T{}
	}

	var ss []string
	if len(n) > 0 {
		ss = strings.SplitN(s, sep, n[0])
	} else {
		ss = strings.Split(s, sep)
	}
	res := make([]T, len(ss))
	ni := make([]uint32, 0, len(ss))
	for i := range ss {
		if v := strings.TrimSpace(ss[i]); v != "" {
			ztype.To(v, &res[i])
		} else {
			ni = append(ni, uint32(i))
		}
	}

	for i := range ni {
		res = append(res[:ni[i]], res[ni[i]+1:]...)
	}
	return res
}

// Join slice to string.
// If value is not empty, the string will be split into value parts.
func Join[T comparable](s []T, sep string) string {
	if len(s) == 0 {
		return ""
	}

	b := zstring.Buffer(len(s))
	for i := 0; i < len(s); i++ {
		v := ztype.ToString(s[i])
		if v == "" {
			continue
		}
		b.WriteString(v)
		if i < len(s)-1 {
			b.WriteString(sep)
		}
	}

	return b.String()
}
