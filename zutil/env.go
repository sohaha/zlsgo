package zutil

import (
	"os"
	"runtime"
	"strings"

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

// Getenv get ENV value by key name
func Getenv(name string, def ...string) string {
	val := os.Getenv(name)
	if val == "" && len(def) > 0 {
		val = def[0]
	}
	return val
}

func GOROOT() string {
	return zfile.RealPath(runtime.GOROOT())
}

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

func loadFile(filename string) error {
	return zfile.ReadLineFile(filename, func(line int, data []byte) error {
		e := strings.Split(zstring.Bytes2String(data), "=")
		var value string
		if len(e) > 1 {
			value = e[1]
		}

		_ = os.Setenv(zstring.TrimSpace(e[0]), strings.TrimSpace(value))
		return nil
	})
}
