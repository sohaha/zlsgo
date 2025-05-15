package zfile

import (
	"errors"
	"os"
	"sync"
)

var (
	// ErrLocked is returned when attempting to lock a file that is already locked
	ErrLocked = errors.New("file is locked")
	// ErrNotLocked is returned when attempting to unlock a file that is not locked
	ErrNotLocked = errors.New("file is not locked")
)

// FileLock provides cross-process file locking capabilities.
// It can be used to synchronize access to resources between multiple processes.
type FileLock struct {
	file *os.File   // The locked file handle
	path string     // Path to the lock file
	mu   sync.Mutex // Mutex for thread-safety within the process
}

// NewFileLock creates a new file lock instance for the specified path.
// The lock is not acquired until Lock() is called.
func NewFileLock(path string) *FileLock {
	return &FileLock{
		path: RealPath(path),
	}
}

// Lock acquires the file lock, blocking other processes from acquiring it.
// If the lock is already held by another process, it returns ErrLocked.
// If the lock is already held by this instance, it returns nil immediately.
func (l *FileLock) Lock() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return nil
	}

	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_RDWR, 0o666)
	if err != nil {
		return err
	}

	err = lockFile(f)
	if err != nil {
		f.Close()
		if err == errLocked {
			return ErrLocked
		}
		return err
	}

	l.file = f
	return nil
}

// Unlock releases the file lock, allowing other processes to acquire it.
// If the lock is not currently held by this instance, it returns ErrNotLocked.
func (l *FileLock) Unlock() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file == nil {
		return ErrNotLocked
	}

	err := unlockFile(l.file)
	if err != nil {
		return err
	}

	err = l.file.Close()
	l.file = nil
	return err
}

// Clean releases the lock if held and removes the lock file from the filesystem.
// This should be called when the lock is no longer needed to clean up resources.
func (l *FileLock) Clean() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		_ = unlockFile(l.file)
		_ = l.file.Close()
		l.file = nil
	}
	return Remove(l.path)
}
