//go:build windows
// +build windows

package zfile

import (
	"os"
	"syscall"
)

// MoveFile moves a file from source to destination path.
// If force is true and destination exists, it will be removed before moving.
// On Windows systems, this uses syscall.MoveFile which provides better support
// for Windows file paths and attributes.
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
