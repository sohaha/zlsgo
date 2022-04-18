//go:build windows
// +build windows

package zfile

import (
	"os"
	"syscall"
)

// MoveFile Move File
func MoveFile(source string, dest string, force ...bool) error {
	source = RealPath(source)
	dest = RealPath(dest)
	if len(force) > 0 && force[0] {
		if exist, _ := PathExist(dest); exist != 0 && source != dest {
			_ = os.RemoveAll(dest)
		}
	}
	from, err := syscall.UTF16PtrFromString(source)
	if err != nil {
		return err
	}
	to, err := syscall.UTF16PtrFromString(dest)
	if err != nil {
		return err
	}
	return syscall.MoveFile(from, to)
}
