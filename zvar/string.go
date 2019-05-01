package zvar

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	// StrPadRight 字符串右边填充
	StrPadRight int = iota
	// StrPadLeft 字符串左边填充
	StrPadLeft
)

// StrPad 字符串填充
func StrPad(raw string, length int, padStr string, padType int) string {
	l := length - StrLen(raw)
	if l <= 0 {
		return raw
	}
	if padType == StrPadRight {
		raw = fmt.Sprintf("%s%s", raw, strings.Repeat(padStr, l))
	} else {
		raw = fmt.Sprintf("%s%s", strings.Repeat(padStr, l), raw)
	}
	return raw
}

// StrLen 字符串长度（中文）
func StrLen(str string) int {
	// strings.Count(str,"")-1
	return utf8.RuneCountInString(str)
}
