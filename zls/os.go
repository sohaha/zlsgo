package zls

import (
	"runtime"
)

// IsWin IsWin
func IsWin() bool {
	return runtime.GOOS == "windows"
}
