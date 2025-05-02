// Package zstring provides String related operations
package zstring

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

type (
	// ru is a pseudorandom number generator
	ru struct {
		x uint32
	}
	PadType uint8
)

const (
	// PadRight Right padding character
	PadRight PadType = iota
	// PadLeft Left padding character
	PadLeft
	// PadSides Two-sided padding characters,If the two sides are not equal, the right side takes precedence.
	PadSides
)

var letterBytes = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// Pad String padding
func Pad(raw string, length int, padStr string, padType PadType) string {
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
	var size, ll, n, nn int
	if len(length) > 0 {
		ll = length[0] + start
	}
	lb := ll == 0
	if start < 0 {
		start = Len(str) + start
	}
	for i := 0; i < len(str); i++ {
		_, size = utf8.DecodeRuneInString(str[nn:])
		if i < start {
			n += size
		} else if lb {
			break
		}
		if !lb && i < ll {
			nn += size
		} else if lb {
			nn += size
		}
	}
	if !lb {
		return str[n:nn]
	}
	return str[n:]
}

// Bytes2String bytes to string
func Bytes2String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// String2Bytes string to bytes
// remark: read only, the structure of runtime changes will be affected, the role of unsafe.Pointer will be changed, and it will also be affected
func String2Bytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

// Ucfirst First letters capitalize
func Ucfirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

// Lcfirst First letters lowercase
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

// TrimBOM TrimBOM
func TrimBOM(fileBytes []byte) []byte {
	trimmedBytes := bytes.Trim(fileBytes, "\xef\xbb\xbf")
	return trimmedBytes
}

// SnakeCaseToCamelCase snakeCase To CamelCase: hello_world => helloWorld
func SnakeCaseToCamelCase(str string, ucfirst bool, delimiter ...string) string {
	if str == "" {
		return ""
	}
	sep := "_"
	if len(delimiter) > 0 {
		sep = delimiter[0]
	}
	slice := strings.Split(str, sep)
	for i := range slice {
		if ucfirst || i > 0 {
			slice[i] = strings.Title(slice[i])
		}
	}
	return strings.Join(slice, "")
}

// CamelCaseToSnakeCase camelCase To SnakeCase helloWorld/HelloWorld => hello_world
func CamelCaseToSnakeCase(str string, delimiter ...string) string {
	if str == "" {
		return ""
	}
	sep := []byte("_")
	if len(delimiter) > 0 {
		sep = []byte(delimiter[0])
	}
	strLen := len(str)
	result := make([]byte, 0, strLen*2)
	j := false
	for i := 0; i < strLen; i++ {
		char := str[i]
		if i > 0 && char >= 'A' && char <= 'Z' && j {
			result = append(result, sep...)
		}
		if char != '_' {
			j = true
		}
		result = append(result, char)
	}
	return strings.ToLower(string(result))
}

// XSSClean clean html tag
func XSSClean(str string) string {
	str, _ = RegexReplaceFunc("<[\\S\\s]+?>", str, strings.ToLower)
	str, _ = RegexReplace("<style[\\S\\s]+?</style>", str, "")
	str, _ = RegexReplace("<script[\\S\\s]+?</script>", str, "")
	str, _ = RegexReplace("<[\\S\\s]+?>", str, "")
	str, _ = RegexReplace("\\s{2,}", str, " ")
	return strings.TrimSpace(str)
}

// TrimLine TrimLine
func TrimLine(s string) string {
	str := strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(s, " "))
	str = strings.Replace(str, " <", "<", -1)
	str = strings.Replace(str, "> ", ">", -1)
	return str
}

var space = [...]uint8{127, 128, 133, 160, 194, 226, 227}

func well(s uint8) bool {
	for i := range space {
		if space[i] == s {
			return true
		}
	}
	return false
}

// TrimSpace TrimSpace
func TrimSpace(s string) string {
	for len(s) > 0 {
		if (s[0] <= 31) || s[0] <= ' ' || well(s[0]) {
			s = s[1:]
			continue
		}
		break
	}
	for len(s) > 0 {
		if s[len(s)-1] <= ' ' || (s[len(s)-1] <= 31) || well(s[len(s)-1]) {
			s = s[:len(s)-1]
			continue
		}
		break
	}
	return s
}

// IsSpace is space character
func IsSpace(r rune) bool {
	switch r {
	case '\t', '\v', '\f', '\r', '\n', ' ', 0x85, 0xA0:
		return true
	}
	return false
}
