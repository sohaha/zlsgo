package zfile

import (
	"io"
	"os"
	"sync"
	"time"
)

// MemoryFile implements an in-memory file that can be periodically flushed to disk.
// It implements the os.FileInfo interface and provides buffered file operations.
type MemoryFile struct {
	stopTiming chan struct{}         // Channel to stop the flush timer
	fbefore    memoryFileFlushBefore // Function to call before flushing
	name       string                // Name of the file on disk
	buffer     byteBuffer            // In-memory buffer for file contents
	timing     int64                 // Auto-flush interval in seconds
	lock       sync.RWMutex          // Lock for thread-safe operations
	stop       bool                  // Flag indicating if the file is closed
}

// MemoryFileOption is a function type for configuring a MemoryFile.
type MemoryFileOption func(*MemoryFile)

// memoryFileFlushBefore is a function type called before flushing to disk.
type memoryFileFlushBefore func(f *MemoryFile) error

// MemoryFileAutoFlush creates an option that configures automatic flushing
// of the memory file to disk at the specified interval in seconds.
func MemoryFileAutoFlush(second int64) func(*MemoryFile) {
	return func(f *MemoryFile) {
		f.timing = second
	}
}

// MemoryFileFlushBefore creates an option that sets a function to be called before
// flushing to disk. Take care to avoid writing to the file in this function to prevent deadlocks.
func MemoryFileFlushBefore(fn memoryFileFlushBefore) func(*MemoryFile) {
	return func(f *MemoryFile) {
		f.fbefore = fn
	}
}

// NewMemoryFile creates a new in-memory file with the specified name and options.
// The file will be flushed to disk when Sync() is called or automatically if configured.
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

// flushLoop is an internal method that periodically flushes the file to disk
// based on the configured timing interval.
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

// SetName changes the name of the file used when flushing to disk.
func (f *MemoryFile) SetName(name string) {
	f.name = name
}

// Bytes returns a copy of the current contents of the memory file.
func (f *MemoryFile) Bytes() []byte {
	f.lock.RLock()
	b := f.buffer.buffer
	f.lock.RUnlock()
	return b
}

// Stat returns file information. The MemoryFile itself implements os.FileInfo.
func (f *MemoryFile) Stat() (os.FileInfo, error) {
	return f, nil
}

// Read reads data from the memory file into the provided buffer.
// It implements the io.Reader interface.
func (f *MemoryFile) Read(buffer []byte) (int, error) {
	f.lock.RLock()
	n, err := f.buffer.Read(buffer)
	f.lock.RUnlock()
	return n, err
}

// Close flushes any pending data to disk and stops the auto-flush timer if active.
// It implements the io.Closer interface.
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

// Sync flushes the memory buffer to disk.
// If the file is empty or if the flush-before function returns an error, no flush occurs.
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

// Write writes data to the memory file.
// It implements the io.Writer interface.
func (f *MemoryFile) Write(buffer []byte) (int, error) {
	f.lock.Lock()
	n, err := f.buffer.Write(buffer)
	f.lock.Unlock()
	return n, err
}

// Seek sets the offset for the next Read or Write operation.
// It implements the io.Seeker interface.
func (f *MemoryFile) Seek(offset int64, whence int) (int64, error) {
	f.lock.RLock()
	n, err := f.buffer.Seek(offset, whence)
	f.lock.RUnlock()
	return n, err
}

// Name returns the name of the file.
// This is part of the os.FileInfo interface implementation.
func (f *MemoryFile) Name() string {
	return f.name
}

// Size returns the size of the file in bytes.
// This is part of the os.FileInfo interface implementation.
func (f *MemoryFile) Size() int64 {
	f.lock.RLock()
	l := int64(f.buffer.Len())
	f.lock.RUnlock()
	return l
}

// Mode returns the file mode bits.
// This is part of the os.FileInfo interface implementation.
func (f *MemoryFile) Mode() os.FileMode {
	return 0o666
}

// ModTime returns the modification time.
// This is part of the os.FileInfo interface implementation.
func (f *MemoryFile) ModTime() time.Time {
	return time.Time{}
}

// IsDir returns whether the file is a directory.
// This is part of the os.FileInfo interface implementation.
func (f *MemoryFile) IsDir() bool {
	return false
}

// Sys returns the underlying data source.
// This is part of the os.FileInfo interface implementation.
func (f *MemoryFile) Sys() interface{} {
	return nil
}

// byteBuffer is an internal structure that implements a seekable byte buffer.
type byteBuffer struct {
	buffer []byte // The actual data
	index  int    // Current read/write position
}

// makeByteBuffer creates a new byte buffer with the given initial content.
func makeByteBuffer(buffer []byte) byteBuffer {
	return byteBuffer{
		buffer: buffer,
		index:  0,
	}
}

// Reset clears the buffer and resets the position to the beginning.
func (bb *byteBuffer) Reset() {
	bb.buffer = bb.buffer[:0]
	bb.index = 0
}

// Len returns the length of the buffer in bytes.
func (bb *byteBuffer) Len() int {
	return len(bb.buffer)
}

// Position returns the current read/write position in the buffer.
func (bb *byteBuffer) Position() int {
	return bb.index
}

// Bytes returns the underlying byte slice of the buffer.
func (bb *byteBuffer) Bytes() []byte {
	return bb.buffer
}

// Read reads data from the buffer into the provided slice.
// It implements the io.Reader interface.
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

// Write writes data to the buffer at the current position.
// It implements the io.Writer interface.
func (bb *byteBuffer) Write(buffer []byte) (int, error) {
	bb.buffer = append(bb.buffer[:bb.index], buffer...)
	bb.index += len(buffer)
	return len(buffer), nil
}

// Seek sets the position for the next Read or Write operation.
// It implements the io.Seeker interface.
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
