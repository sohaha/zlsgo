// Package zstring provides string manipulation utilities.
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
	// ru is a pseudorandom number generator used for string operations
	// that require randomization.
	ru struct {
		x uint32
	}
	// PadType defines the padding strategy for string padding operations.
	PadType uint8
)

const (
	// PadRight indicates padding should be added to the right side of the string.
	PadRight PadType = iota
	// PadLeft indicates padding should be added to the left side of the string.
	PadLeft
	// PadSides indicates padding should be added to both sides of the string.
	// If the padding cannot be distributed equally, the right side receives the extra character.
	PadSides
)

var letterBytes = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// Pad adds padding characters to a string to reach the specified length.
// The padType parameter controls where padding is added (left, right, or both sides).
// If the string is already longer than the specified length, it is returned unchanged.
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

// Len returns the number of characters (runes) in a UTF-8 encoded string.
// This differs from len(string) which returns the number of bytes.
func Len(str string) int {
	// strings.Count(str,"")-1
	return utf8.RuneCountInString(str)
}

// Substr extracts a substring from a UTF-8 encoded string.
// The start parameter specifies the position of the first character (can be negative to count from the end).
// The optional length parameter specifies how many characters to include in the result.
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

// Bytes2String converts a byte slice to a string without memory allocation.
// Note: This uses unsafe.Pointer and the returned string must not be modified.
func Bytes2String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// String2Bytes converts a string to a byte slice without memory allocation.
// Note: This uses unsafe.Pointer and the returned byte slice must be treated as read-only.
// Modifying the returned slice may cause undefined behavior as it shares memory with the original string.
func String2Bytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

// Ucfirst capitalizes the first letter of a string, leaving the rest unchanged.
// Returns an empty string if the input is empty.
func Ucfirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

// Lcfirst converts the first letter of a string to lowercase, leaving the rest unchanged.
// Returns an empty string if the input is empty.
func Lcfirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

// IsUcfirst checks if the first letter of a string is uppercase.
// Returns false if the string is empty or the first character is not a letter.
func IsUcfirst(str string) bool {
	b := String2Bytes(str)
	if b[0] >= byte('A') && b[0] <= byte('Z') {
		return true
	}
	return false
}

// IsLcfirst checks if the first letter of a string is lowercase.
// Returns false if the string is empty or the first character is not a letter.
func IsLcfirst(str string) bool {
	b := String2Bytes(str)
	if b[0] >= byte('a') && b[0] <= byte('z') {
		return true
	}
	return false
}

// TrimBOM removes the UTF-8 Byte Order Mark (BOM) from the beginning of a byte slice if present.
// The BOM is the byte sequence 0xEF,0xBB,0xBF that sometimes appears at the start of UTF-8 encoded files.
func TrimBOM(fileBytes []byte) []byte {
	trimmedBytes := bytes.Trim(fileBytes, "\xef\xbb\xbf")
	return trimmedBytes
}

// SnakeCaseToCamelCase converts a snake_case string to camelCase or PascalCase.
// Example: "hello_world" becomes "helloWorld" (or "HelloWorld" if ucfirst is true).
// The optional delimiter parameter specifies the separator character (default is "_").
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

// CamelCaseToSnakeCase converts a camelCase or PascalCase string to snake_case.
// Example: "helloWorld" or "HelloWorld" becomes "hello_world".
// The optional delimiter parameter specifies the separator character (default is "_").
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

// XSSClean removes HTML and JavaScript tags from a string to prevent XSS attacks.
// It removes style tags, script tags, and all other HTML tags, then normalizes whitespace.
func XSSClean(str string) string {
	str, _ = RegexReplaceFunc("<[\\S\\s]+?>", str, strings.ToLower)
	str, _ = RegexReplace("<style[\\S\\s]+?</style>", str, "")
	str, _ = RegexReplace("<script[\\S\\s]+?</script>", str, "")
	str, _ = RegexReplace("<[\\S\\s]+?>", str, "")
	str, _ = RegexReplace("\\s{2,}", str, " ")
	return strings.TrimSpace(str)
}

// TrimLine removes leading and trailing whitespace from each line in a string,
// and removes empty lines. It preserves the newline characters between non-empty lines.
func TrimLine(s string) string {
	str := strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(s, " "))
	str = strings.Replace(str, " <", "<", -1)
	str = strings.Replace(str, "> ", ">", -1)
	return str
}

var space = [...]uint8{127, 128, 133, 160, 194, 226, 227}

// well checks if a byte is one of the special whitespace characters defined in the space array.
// Used internally by TrimSpace to handle additional Unicode whitespace characters.
func well(s uint8) bool {
	for i := range space {
		if space[i] == s {
			return true
		}
	}
	return false
}

// TrimSpace removes all leading and trailing whitespace from a string.
// Unlike the standard strings.TrimSpace, this function handles additional Unicode whitespace characters.
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

// IsSpace checks if a rune is a whitespace character.
// This includes standard ASCII whitespace and additional Unicode whitespace characters.
func IsSpace(r rune) bool {
	switch r {
	case '\t', '\v', '\f', '\r', '\n', ' ', 0x85, 0xA0:
		return true
	}
	return false
}
