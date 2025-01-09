//go:build !windows
// +build !windows

package zfile

import (
	"os"
	"syscall"
)

var errLocked = syscall.EWOULDBLOCK

func lockFile(f *os.File) error {
	err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		if err == syscall.EWOULDBLOCK {
			return errLocked
		}
		return &os.PathError{Op: "lock", Path: f.Name(), Err: err}
	}
	return nil
}

func unlockFile(f *os.File) error {
	err := syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	if err != nil {
		return &os.PathError{Op: "unlock", Path: f.Name(), Err: err}
	}
	return nil
}
