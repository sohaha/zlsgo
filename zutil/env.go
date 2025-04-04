package zutil

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"unicode"

	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zstring"
)

func GetOs() string {
	return runtime.GOOS
}

// IsWin system. linux windows darwin
func IsWin() bool {
	return GetOs() == "windows"
}

// IsMac system
func IsMac() bool {
	return GetOs() == "darwin"
}

// IsLinux system
func IsLinux() bool {
	return GetOs() == "linux"
}

// Is32BitArch is 32-bit system
func Is32BitArch() bool {
	return strconv.IntSize == 32
}

// Getenv get ENV value by key name
func Getenv(name string, def ...string) string {
	val := os.Getenv(name)
	if val == "" && len(def) > 0 {
		val = def[0]
	}
	return val
}

// GOROOT return go root path
func GOROOT() string {
	return zfile.RealPath(runtime.GOROOT())
}

// Loadenv load env from file
func Loadenv(filenames ...string) (err error) {
	filenames = filenamesOrDefault(filenames)

	for _, filename := range filenames {
		err = loadFile(filename)
		if err != nil {
			return
		}
	}
	return
}

func filenamesOrDefault(filenames []string) []string {
	if len(filenames) == 0 {
		return []string{".env"}
	}
	return filenames
}

func locateKeyName(src []byte) (key string, cutset []byte, err error) {
	src = bytes.TrimLeftFunc(src, isSpace)
	offset := 0
loop:
	for i, char := range src {
		rchar := rune(char)
		if isSpace(rchar) {
			continue
		}

		switch char {
		case '=', ':':
			key = string(src[0:i])
			offset = i + 1
			break loop
		case '_':
		default:
			if unicode.IsLetter(rchar) || unicode.IsNumber(rchar) || rchar == '.' {
				continue
			}

			return "", nil, fmt.Errorf(`unexpected character %q in variable name near %q`, char, src)
		}
	}

	if len(src) == 0 {
		return "", nil, errors.New("zero length string")
	}

	key = strings.TrimRightFunc(key, unicode.IsSpace)
	cutset = bytes.TrimLeftFunc(src[offset:], isSpace)
	return key, cutset, nil
}

func isSpace(r rune) bool {
	switch r {
	case '\t', '\v', '\f', '\r', ' ', 0x85, 0xA0:
		return true
	}
	return false
}

func loadFile(filename string) error {
	return zfile.ReadLineFile(filename, func(line int, data []byte) error {
		key, value, err := locateKeyName(data)
		if err != nil {
			return nil
		}

		value = bytes.TrimSpace(value)
		if len(value) > 0 && (value[0] == '"' || value[0] == '\'' && value[0] == value[len(value)-1]) {
			value = value[1 : len(value)-1]
		}

		_ = os.Setenv(key, zstring.Bytes2String(value))
		return nil
	})
}
