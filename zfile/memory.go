package zfile

import (
	"io"
	"os"
	"sync"
	"time"
)

type MemoryFile struct {
	stopTiming chan struct{}
	fbefore    memoryFileFlushBefore
	name       string
	buffer     byteBuffer
	timing     int64
	lock       sync.RWMutex
	stop       bool
}

type MemoryFileOption func(*MemoryFile)
type memoryFileFlushBefore func(f *MemoryFile) error

func MemoryFileAutoFlush(second int64) func(*MemoryFile) {
	return func(f *MemoryFile) {
		f.timing = second
	}
}

// MemoryFileFlushBefore is a function that will be called before flush to disk,take care to avoid writing to prevent deadlocks
func MemoryFileFlushBefore(fn memoryFileFlushBefore) func(*MemoryFile) {
	return func(f *MemoryFile) {
		f.fbefore = fn
	}
}

func NewMemoryFile(name string, opt ...MemoryFileOption) *MemoryFile {
	f := &MemoryFile{name: name, buffer: makeByteBuffer([]byte{})}
	for _, o := range opt {
		o(f)
	}
	f.stopTiming = make(chan struct{})
	if f.timing > 0 {
		go f.flushLoop()
	}
	return f
}

func (f *MemoryFile) flushLoop() {
	ticker := time.NewTicker(time.Second * time.Duration(f.timing))
	for {
		select {
		case <-f.stopTiming:
			return
		case <-ticker.C:
			_ = f.Sync()
		}
	}
}

func (f *MemoryFile) SetName(name string) {
	f.name = name
}

func (f *MemoryFile) Bytes() []byte {
	f.lock.RLock()
	b := f.buffer.buffer
	f.lock.RUnlock()
	return b
}

func (f *MemoryFile) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *MemoryFile) Read(buffer []byte) (int, error) {
	f.lock.RLock()
	n, err := f.buffer.Read(buffer)
	f.lock.RUnlock()
	return n, err
}

func (f *MemoryFile) Close() error {
	f.lock.Lock()
	if f.stop {
		return nil
	}
	f.stop = true
	f.lock.Unlock()
	f.stopTiming <- struct{}{}
	return f.Sync()
}

func (f *MemoryFile) Sync() error {
	if f.Size() == 0 {
		return nil
	}
	if f.fbefore != nil {
		err := f.fbefore(f)
		if err != nil {
			return err
		}
	}
	f.lock.Lock()
	b := f.buffer.buffer
	f.buffer.Reset()
	f.lock.Unlock()
	return WriteFile(f.name, b, true)
}

func (f *MemoryFile) Write(buffer []byte) (int, error) {
	f.lock.Lock()
	n, err := f.buffer.Write(buffer)
	f.lock.Unlock()
	return n, err
}

func (f *MemoryFile) Seek(offset int64, whence int) (int64, error) {
	f.lock.RLock()
	n, err := f.buffer.Seek(offset, whence)
	f.lock.RUnlock()
	return n, err
}

func (f *MemoryFile) Name() string {
	return f.name
}

func (f *MemoryFile) Size() int64 {
	f.lock.RLock()
	l := int64(f.buffer.Len())
	f.lock.RUnlock()
	return l
}

func (f *MemoryFile) Mode() os.FileMode {
	return 0666
}

func (f *MemoryFile) ModTime() time.Time {
	return time.Time{}
}

func (f *MemoryFile) IsDir() bool {
	return false
}

func (f *MemoryFile) Sys() interface{} {
	return nil
}

type byteBuffer struct {
	buffer []byte
	index  int
}

func makeByteBuffer(buffer []byte) byteBuffer {
	return byteBuffer{
		buffer: buffer,
		index:  0,
	}
}

func (bb *byteBuffer) Reset() {
	bb.buffer = bb.buffer[:0]
	bb.index = 0
}

func (bb *byteBuffer) Len() int {
	return len(bb.buffer)
}

func (bb *byteBuffer) Position() int {
	return bb.index
}

func (bb *byteBuffer) Bytes() []byte {
	return bb.buffer
}

func (bb *byteBuffer) Read(buffer []byte) (int, error) {
	if len(buffer) == 0 {
		return 0, nil
	}

	if bb.index >= bb.Len() {
		return 0, io.EOF
	}

	last := copy(buffer, bb.buffer[bb.index:])
	bb.index += last
	return last, nil
}

func (bb *byteBuffer) Write(buffer []byte) (int, error) {
	bb.buffer = append(bb.buffer[:bb.index], buffer...)
	bb.index += len(buffer)
	return len(buffer), nil
}

func (bb *byteBuffer) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	default:
	case io.SeekStart:
		bb.index = int(offset)
	case io.SeekCurrent:
		bb.index += int(offset)
	case io.SeekEnd:
		bb.index = bb.Len() - 1 - int(offset)
	}
	return int64(bb.index), nil
}
