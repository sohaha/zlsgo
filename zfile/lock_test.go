package zfile

import (
	"sync"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
)

func TestFileLock(t *testing.T) {
	tt := zlsgo.NewTest(t)
	lockFile := TmpPath() + "/test.lock"

	defer Remove(lockFile)

	lock1 := NewFileLock(lockFile)
	lock2 := NewFileLock(lockFile)

	err := lock1.Lock()
	tt.NoError(err)

	err = lock1.Lock()
	tt.NoError(err)
	err = lock2.Lock()
	tt.Equal(ErrLocked, err)

	err = lock1.Unlock()
	tt.NoError(err)

	err = lock1.Unlock()
	tt.Equal(ErrNotLocked, err)

	err = lock2.Lock()
	tt.NoError(err)
	err = lock2.Unlock()
	tt.NoError(err)

	done := make(chan bool)
	go func() {
		lock := NewFileLock(lockFile)
		err := lock.Lock()
		tt.NoError(err)
		time.Sleep(100 * time.Millisecond)
		err = lock.Unlock()
		tt.NoError(err)
		done <- true
	}()

	time.Sleep(50 * time.Millisecond)
	lock := NewFileLock(lockFile)
	err = lock.Lock()
	tt.Equal(ErrLocked, err)
	<-done
	err = lock.Lock()
	tt.NoError(err)
	err = lock.Unlock()
	tt.NoError(err)
}

func TestFileLockCleanup(t *testing.T) {
	tt := zlsgo.NewTest(t)
	lockFile := TmpPath() + "/cleanup.lock"

	lock := NewFileLock(lockFile)
	err := lock.Lock()
	tt.NoError(err)

	err = lock.Clean()
	tt.NoError(err)
	tt.EqualTrue(!FileExist(lockFile))

	err = lock.Clean()
	tt.NotNil(err)
}

func TestFileLockConcurrent(t *testing.T) {
	tt := zlsgo.NewTest(t)
	lockFile := TmpPath() + "/concurrent.lock"
	defer Remove(lockFile)

	const goroutines = 10
	var wg sync.WaitGroup
	errors := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lock := NewFileLock(lockFile)
			if err := lock.Lock(); err != nil {
				if err != ErrLocked {
					errors <- err
				}
				return
			}
			time.Sleep(10 * time.Millisecond)
			if err := lock.Unlock(); err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		tt.NoError(err)
	}
}
