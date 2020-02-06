/*
 * @Author: seekwe
 * @Date:   2019-05-09 12:44:23
 * @Last Modified by:   seekwe
 * @Last Modified time: 2020-01-30 18:35:06
 */

// Package zstring provides String related operations
package zstring

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

type sliceT struct {
	array unsafe.Pointer
	len   int
	cap   int
}

type stringStruct struct {
	str unsafe.Pointer
	len int
}

const (
	// PadRight Right padding character
	PadRight int = iota
	// PadLeft Left padding character
	PadLeft
	// PadSides Two-sided padding characters,If the two sides are not equal, the right side takes precedence.
	PadSides
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxMask = 1<<6 - 1 // All 1-bits, as many as 6
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

// Len string length (utf8)
func Len(str string) int {
	// strings.Count(str,"")-1
	return utf8.RuneCountInString(str)
}

// Substr returns part of a string
func Substr(str string, start int, length ...int) string {
	s := []rune(str)
	sl := len(s)
	if start < 0 {
		start = sl + start
	}

	if len(length) > 0 {
		ll := length[0]
		if ll < 0 {
			sl = sl + ll
		} else {
			sl = ll + start
		}
	}
	return string(s[start:sl])
}

// Bytes2String bytes to string
func Bytes2String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// String2Bytes string to bytes
func String2Bytes(s string) []byte {
	str := (*stringStruct)(unsafe.Pointer(&s))
	ret := sliceT{array: str.str, len: str.len, cap: str.len}
	return *(*[]byte)(unsafe.Pointer(&ret))
}

// Ucfirst Ucfirst
func Ucfirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

// Lcfirst Lcfirst
func Lcfirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

// IsUcfirst tests whether the given byte b is in upper case
func IsUcfirst(str string) bool {
	b := String2Bytes(str)
	if b[0] >= byte('A') && b[0] <= byte('Z') {
		return true
	}
	return false
}

// IsLcfirst tests whether the given byte b is in lower case
func IsLcfirst(str string) bool {
	b := String2Bytes(str)
	if b[0] >= byte('a') && b[0] <= byte('z') {
		return true
	}
	return false
}

func TrimBOM(fileBytes []byte) []byte {
	trimmedBytes := bytes.Trim(fileBytes, "\xef\xbb\xbf")
	return trimmedBytes
}
