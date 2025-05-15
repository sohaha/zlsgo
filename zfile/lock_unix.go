//go:build !windows
// +build !windows

package zfile

import (
	"os"
	"syscall"
)

// errLocked is the error returned when a file is already locked by another process.
// On Unix systems, this corresponds to syscall.EWOULDBLOCK.
var errLocked = syscall.EWOULDBLOCK

// lockFile acquires an exclusive, non-blocking lock on the given file.
// On Unix systems, this uses syscall.Flock with LOCK_EX|LOCK_NB flags.
// Returns errLocked if the file is already locked by another process.
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

// unlockFile releases a lock previously acquired with lockFile.
// On Unix systems, this uses syscall.Flock with LOCK_UN flag.
func unlockFile(f *os.File) error {
	err := syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	if err != nil {
		return &os.PathError{Op: "unlock", Path: f.Name(), Err: err}
	}
	return nil
}
