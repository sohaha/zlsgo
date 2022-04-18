package zutil

import (
	"os"
	"runtime"

	"github.com/sohaha/zlsgo/zfile"
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
