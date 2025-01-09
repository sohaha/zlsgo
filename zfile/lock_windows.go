//go:build windows
// +build windows

package zfile

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	lockfileExclusiveLock   = 0x2
	lockfileFailImmediately = 0x1
	errorLockViolation      = 0x21
)

var errLocked = syscall.EWOULDBLOCK

var (
	kernel32DLL      = syscall.NewLazyDLL("kernel32.dll")
	procLockFileEx   = kernel32DLL.NewProc("LockFileEx")
	procUnlockFileEx = kernel32DLL.NewProc("UnlockFileEx")
)

func lockFile(f *os.File) error {
	h := syscall.Handle(f.Fd())
	var ol syscall.Overlapped

	ol.Offset = 0
	ol.OffsetHigh = 0

	r1, _, err := procLockFileEx.Call(
		uintptr(h),
		uintptr(lockfileExclusiveLock|lockfileFailImmediately),
		0,
		1,
		0,
		uintptr(unsafe.Pointer(&ol)),
	)

	if r1 == 0 {
		if e, ok := err.(syscall.Errno); ok && e == errorLockViolation {
			return errLocked
		}
		return &os.PathError{Op: "lock", Path: f.Name(), Err: err}
	}
	return nil
}

func unlockFile(f *os.File) error {
	h := syscall.Handle(f.Fd())
	var ol syscall.Overlapped

	ol.Offset = 0
	ol.OffsetHigh = 0

	r1, _, err := procUnlockFileEx.Call(
		uintptr(h),
		0,
		1,
		0,
		uintptr(unsafe.Pointer(&ol)),
	)

	if r1 == 0 {
		return &os.PathError{Op: "unlock", Path: f.Name(), Err: err}
	}
	return nil
}
