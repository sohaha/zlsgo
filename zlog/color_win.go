// +build windows

package zlog

import (
	"syscall"
)

var (
	winEnable          bool
	procSetConsoleMode *syscall.LazyProc
)

func init() {
	if supportColor || isMsystem {
		return
	}
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")

	winEnable = tryApplyOnCONOUT()
	if !winEnable {
		winEnable = tryApplyStdout()
	}
}

func tryApplyOnCONOUT() bool {
	outHandle, err := syscall.Open("CONOUT$", syscall.O_RDWR, 0)
	if err != nil {
		return false
	}

	err = EnableTerminalProcessing(outHandle, true)
	if err != nil {
		return false
	}

	return true
}

func tryApplyStdout() bool {
	err := EnableTerminalProcessing(syscall.Stdout, true)
	if err != nil {
		return false
	}

	return true
}

func EnableTerminalProcessing(stream syscall.Handle, enable bool) error {
	var mode uint32
	err := syscall.GetConsoleMode(stream, &mode)
	if err != nil {
		return err
	}

	if enable {
		mode |= 0x4
	} else {
		mode &^= 0x4
	}

	ret, _, err := procSetConsoleMode.Call(uintptr(stream), uintptr(mode))
	if ret == 0 {
		return err
	}

	return nil
}

// IsSupportColor IsSupportColor
func IsSupportColor() bool {
	return supportColor || winEnable
}
