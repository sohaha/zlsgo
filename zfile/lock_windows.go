//go:build windows
// +build windows

package zfile

import (
	"os"
	"syscall"
	"unsafe"
)

// Windows-specific constants for file locking operations
const (
	// lockfileExclusiveLock requests an exclusive lock
	lockfileExclusiveLock = 0x2
	// lockfileFailImmediately returns immediately if the lock cannot be acquired
	lockfileFailImmediately = 0x1
	// errorLockViolation is the error code returned when a lock is already held
	errorLockViolation = 0x21
)

// errLocked is the error returned when a file is already locked by another process.
// For consistency with Unix implementations, we use syscall.EWOULDBLOCK.
var errLocked = syscall.EWOULDBLOCK

// Windows API function references for file locking
var (
	// kernel32DLL is a reference to the Windows kernel32.dll library
	kernel32DLL = syscall.NewLazyDLL("kernel32.dll")
	// procLockFileEx is a reference to the LockFileEx function in kernel32.dll
	procLockFileEx = kernel32DLL.NewProc("LockFileEx")
	// procUnlockFileEx is a reference to the UnlockFileEx function in kernel32.dll
	procUnlockFileEx = kernel32DLL.NewProc("UnlockFileEx")
)

// lockFile acquires an exclusive, non-blocking lock on the given file.
// On Windows systems, this uses the LockFileEx Win32 API function.
// Returns errLocked if the file is already locked by another process.
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

// unlockFile releases a lock previously acquired with lockFile.
// On Windows systems, this uses the UnlockFileEx Win32 API function.
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
