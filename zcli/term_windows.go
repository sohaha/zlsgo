//go:build windows

package zcli

import (
	"os"
	"syscall"
	"unsafe"
)

type coord struct {
	X int16
	Y int16
}

type smallRect struct {
	Left   int16
	Top    int16
	Right  int16
	Bottom int16
}

type consoleScreenBufferInfo struct {
	Size              coord
	CursorPosition    coord
	Attributes        uint16
	Window            smallRect
	MaximumWindowSize coord
}

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
)

func fileIsTerminal(file *os.File) bool {
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func fileTerminalWidth(file *os.File) (int, bool) {
	if !fileIsTerminal(file) {
		return 0, false
	}

	var infoBuf consoleScreenBufferInfo
	r1, _, _ := procGetConsoleScreenBufferInfo.Call(file.Fd(), uintptr(unsafe.Pointer(&infoBuf)))
	if r1 == 0 {
		return 0, false
	}
	width := int(infoBuf.Window.Right-infoBuf.Window.Left) + 1
	if width <= 0 {
		return 0, false
	}
	return width, true
}
