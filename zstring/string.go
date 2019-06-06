/*
 * @Author: seekwe
 * @Date:   2019-05-09 12:44:23
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-06-06 16:11:30
 */

// Package zstring provides String related operations
package zstring

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	// PadRight Right padding character
	PadRight int = iota
	// PadLeft Left padding character
	PadLeft
	// PadSides Two-sided padding characters,If the two sides are not equal, the right side takes precedence.
	PadSides
)

// Pad String padding
func Pad(raw string, length int, padStr string, padType int) string {
	l := length - Len(raw)
	if l <= 0 {
		return raw
	}
	if padType == PadRight {
		raw = fmt.Sprintf("%s%s", raw, strings.Repeat(padStr, l))
	} else if padType == PadLeft {
		raw = fmt.Sprintf("%s%s", strings.Repeat(padStr, l), raw)
	} else {
		left := 0
		right := 0
		if l > 1 {
			left = l / 2
			right = (l / 2) + (l % 2)
		}

		raw = fmt.Sprintf("%s%s%s", strings.Repeat(padStr, left), raw, strings.Repeat(padStr, right))
	}
	return raw
}

// Len String length (Chinese)
func Len(str string) int {
	// strings.Count(str,"")-1
	return utf8.RuneCountInString(str)
}

// Substr substr
func Substr(str string, start int, length ...int) string {
	s := []rune(str)
	var l int
	if len(length) > 0 {
		l = length[0] + start
	} else {
		l = len(s)
	}
	return string(s[start:l])
}
