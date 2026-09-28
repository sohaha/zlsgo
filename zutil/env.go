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

// GetOs returns the current operating system name as reported by the Go runtime.
// Possible values include "windows", "darwin" (macOS), "linux", etc.
func GetOs() string {
	return runtime.GOOS
}

// IsWin checks if the current operating system is Windows.
func IsWin() bool {
	return GetOs() == "windows"
}

// IsMac checks if the current operating system is macOS (darwin).
func IsMac() bool {
	return GetOs() == "darwin"
}

// IsLinux checks if the current operating system is Linux.
func IsLinux() bool {
	return GetOs() == "linux"
}

// Is32BitArch checks if the current architecture is 32-bit.
func Is32BitArch() bool {
	return strconv.IntSize == 32
}

// Getenv retrieves the value of an environment variable by its name.
// If the environment variable is not set and a default value is provided,
// the default value will be returned. An empty string is considered a valid value.
func Getenv(name string, def ...string) string {
	val, exists := os.LookupEnv(name)
	if !exists && len(def) > 0 {
		return def[0]
	}
	return val
}

// GOROOT returns the absolute path to the Go root directory.
// This is equivalent to the GOROOT environment variable but is determined
// by the Go runtime rather than the environment.
func GOROOT() string {
	return zfile.RealPath(runtime.GOROOT())
}

// Loadenv loads environment variables from one or more .env files.
// If no filenames are provided, it defaults to loading from a file named ".env".
//
// Variables already present in the process environment are never overwritten,
// so the real environment always takes precedence over .env files (this matches
// the standard dotenv semantics). When several files are given, later files
// override earlier ones. A value defined in a .env file becomes visible to the
// process environment only when it is not already set.
//
// The file format follows the standard .env format with KEY=VALUE pairs.
// Relative filenames are resolved against zfile.ProjectPath (the executable
// directory by default); assign zfile.ProjectPath to change the lookup base.
func Loadenv(filenames ...string) (err error) {
	filenames = filenamesOrDefault(filenames)

	vars := make(map[string]string)
	for _, filename := range filenames {
		err = loadFile(filename, vars)
		if err != nil {
			return
		}
	}

	for key, value := range vars {
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, value)
	}
	return
}

// filenamesOrDefault returns the provided filenames or a default filename (".env")
// if none were provided. This is an internal helper function for Loadenv.
func filenamesOrDefault(filenames []string) []string {
	if len(filenames) == 0 {
		return []string{".env"}
	}
	return filenames
}

// locateKeyName parses a line from an .env file to extract the key name and value.
// This is an internal helper function for loadFile.
func locateKeyName(src []byte) (key string, cutset []byte, err error) {
	src = bytes.TrimLeftFunc(src, zstring.IsSpace)
	offset := 0
loop:
	for i, char := range src {
		rchar := rune(char)
		if zstring.IsSpace(rchar) {
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
	cutset = bytes.TrimLeftFunc(src[offset:], zstring.IsSpace)
	return key, cutset, nil
}

// loadFile parses KEY=VALUE pairs from a file into vars. Key/value pairs are not
// applied to the process environment here: Loadenv applies them afterwards so
// that real environment variables keep precedence and later files win.
// This is an internal helper function for Loadenv.
func loadFile(filename string, vars map[string]string) error {
	return zfile.ReadLineFile(filename, func(line int, data []byte) error {
		key, value, err := locateKeyName(data)
		if err != nil {
			return nil
		}

		value = bytes.TrimSpace(value)
		if len(value) > 1 && (value[0] == '"' || value[0] == '\'') && value[0] == value[len(value)-1] {
			value = value[1 : len(value)-1]
		}

		vars[key] = zstring.Bytes2String(value)
		return nil
	})
}
