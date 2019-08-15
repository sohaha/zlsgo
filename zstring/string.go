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

var src = rand.NewSource(time.Now().UnixNano())

func Rand(n int) string {
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), 10; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), 10
		}
		b[i] = letterBytes[int(cache&letterIdxMask)%len(letterBytes)]
		i--
		cache >>= 6
		remain--
	}
	return Bytes2String(b)
}

func Bytes2String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func String2Bytes(s *string) []byte {
	return *(*[]byte)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(s))))
}
