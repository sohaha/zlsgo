// +build windows

package zlog

import (
	"syscall"
)

var (
	winEnable bool
	kernel32  *syscall.LazyDLL
	proc      *syscall.LazyProc
)

func init() {
	if isSupportColor() {
		return
	}
	winEnable = false
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	proc = kernel32.NewProc("SetConsoleTextAttribute")
}

// IsSupportColor IsSupportColor
func IsSupportColor() bool {
	return !DisableColor && winEnable
}

func setColor(i int) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("SetConsoleTextAttribute")
	_, _, _ = proc.Call(uintptr(syscall.Stdout), uintptr(i))
}

func resetColor() {
	handle, _, _ := proc.Call(uintptr(syscall.Stdout), uintptr(7))
	CloseHandle := kernel32.NewProc("CloseHandle")
	_, _, _ = CloseHandle.Call(handle)
}
