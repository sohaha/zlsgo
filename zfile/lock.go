package zfile

import (
	"errors"
	"os"
	"sync"
)

var (
	ErrLocked    = errors.New("file is locked")
	ErrNotLocked = errors.New("file is not locked")
)

type FileLock struct {
	file *os.File
	path string
	mu   sync.Mutex
}

// NewFileLock creates a new file lock instance
func NewFileLock(path string) *FileLock {
	return &FileLock{
		path: RealPath(path),
	}
}

// Lock acquires the file lock
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

// Unlock releases the file lock
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

// Clean removes the lock file if it exists
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
