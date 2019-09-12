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
	"math/rand"
	"reflect"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

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

// Rand rand string
func Rand(n int, ostr ...string) string {
	var src = rand.NewSource(time.Now().UnixNano())
	var s string
	b := make([]byte, n)
	if len(ostr) > 0 {
		s = ostr[0]
	} else {
		s = letterBytes
	}
	for i, cache, remain := n-1, src.Int63(), 10; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), 10
		}
		b[i] = s[int(cache&letterIdxMask)%len(s)]
		i--
		cache >>= 6
		remain--
	}
	return Bytes2String(b)
}

// RandomInt randomInteger
func RandomInt(num int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(num)
}

// Bytes2String bytes to string
func Bytes2String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// String2Bytes string to bytes
func String2Bytes(s *string) []byte {
	return *(*[]byte)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(s))))
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
func IsUcfirst(str *string) bool {
	b := String2Bytes(str)
	if b[0] >= byte('A') && b[0] <= byte('Z') {
		return true
	}
	return false
}

// IsLcfirst tests whether the given byte b is in lower case
func IsLcfirst(str *string) bool {
	b := String2Bytes(str)
	if b[0] >= byte('a') && b[0] <= byte('z') {
		return true
	}
	return false
}
