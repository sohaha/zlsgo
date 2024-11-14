package zarray

import (
	"strings"

	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztype"
)

// Slice converts a string to a slice.
// If n is not empty, the string will be split into n parts.
func Slice[T comparable](s string, n ...int) []T {
	if s == "" {
		return []T{}
	}

	var ss []string
	if len(n) > 0 {
		ss = strings.SplitN(s, ",", n[0])
	} else {
		ss = strings.Split(s, ",")
	}
	res := make([]T, len(ss))
	for i := range ss {
		ztype.To(zstring.TrimSpace(ss[i]), &res[i])
	}

	return res
}

// Join slice to string.
// If n is not empty, the string will be split into n parts.
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
