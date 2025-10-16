package zsync

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zutil"
)

func TestSeqLock(t *testing.T) {
	tt := zlsgo.NewTest(t)

	t.Run("Basic", func(t *testing.T) {
		s := NewSeqLock[string]()

		s.Write("hello")
		v, ok := s.Read()
		tt.EqualTrue(ok)
		tt.Equal("hello", v)

		s.Write("123")
		v, ok = s.Read()
		tt.EqualTrue(ok)
		tt.Equal("123", v)
	})

	t.Run("StructData", func(t *testing.T) {
		type Data struct {
			Name  string
			Value int
		}

		s := NewSeqLock[Data]()
		data1 := Data{Name: "test", Value: 100}
		s.Write(data1)

		v, ok := s.Read()
		tt.EqualTrue(ok)
		tt.Equal("test", v.Name)
		tt.Equal(100, v.Value)
	})

	t.Run("ConcurrentReadWrite", func(t *testing.T) {
		s := NewSeqLock[int]()
		const writers = 5
		const readers = 20
		const operations = 1000

		var wg sync.WaitGroup

		for i := 0; i < writers; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < operations; j++ {
					s.Write(id*operations + j)
				}
			}(i)
		}

		readCount := 0
		var mu sync.Mutex
		for i := 0; i < readers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < operations; j++ {
					_, ok := s.Read()
					if ok {
						mu.Lock()
						readCount++
						mu.Unlock()
					}
				}
			}()
		}

		wg.Wait()

		tt.EqualTrue(readCount > 0)
	})

	t.Run("HighContentionReads", func(t *testing.T) {
		s := NewSeqLock[int]()
		const readers = 100
		const readsPerReader = 100

		s.Write(42)

		var wg sync.WaitGroup
		successfulReads := 0
		var mu sync.Mutex

		for i := 0; i < readers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < readsPerReader; j++ {
					v, ok := s.Read()
					if ok && v == 42 {
						mu.Lock()
						successfulReads++
						mu.Unlock()
					}
				}
			}()
		}

		wg.Wait()

		tt.Equal(readers*readsPerReader, successfulReads)
	})
}

type benchData struct {
	A int64
	B int64
	C int64
}

var benchSink int64

func BenchmarkSeqLock_Read(b *testing.B) {
	s := NewSeqLock[*benchData]()
	s.Write(&benchData{A: 1, B: 2, C: 3})

	b.ReportAllocs()
	b.SetParallelism(4)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v, ok := s.Read()
			if !ok {
				continue
			}
			atomic.AddInt64(&benchSink, v.A)
			runtime.KeepAlive(v)
		}
	})
}

func BenchmarkPointer_Read(b *testing.B) {
	p := zutil.NewPointer(unsafe.Pointer(&benchData{A: 1, B: 2, C: 3}))

	b.ReportAllocs()
	b.SetParallelism(4)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			d := (*benchData)(p.Load())
			atomic.AddInt64(&benchSink, d.A)
			runtime.KeepAlive(d)
		}
	})
}

func BenchmarkSeqLock_Write(b *testing.B) {
	s := NewSeqLock[*benchData]()
	var x int64

	b.ReportAllocs()
	b.SetParallelism(4)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v := atomic.AddInt64(&x, 1)
			s.Write(&benchData{A: v, B: v, C: v})
		}
	})
}

func BenchmarkPointer_Write(b *testing.B) {
	p := zutil.NewPointer(nil)
	var x int64

	b.ReportAllocs()
	b.SetParallelism(4)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v := atomic.AddInt64(&x, 1)
			d := &benchData{A: v, B: v, C: v}
			p.Store(unsafe.Pointer(d))
		}
	})
}

func BenchmarkSeqLock_Mixed10(b *testing.B) { // ~10% writes
	runMixedSeqLock(b, 10)
}

func BenchmarkPointer_Mixed10(b *testing.B) { // ~10% writes
	runMixedPointer(b, 10)
}

func runMixedSeqLock(b *testing.B, writeEvery int) {
	s := NewSeqLock[*benchData]()
	s.Write(&benchData{})
	var ctr int64

	b.ReportAllocs()
	b.SetParallelism(4)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			if writeEvery > 0 && i%writeEvery == 0 {
				v := atomic.AddInt64(&ctr, 1)
				s.Write(&benchData{A: v, B: v, C: v})
			} else {
				v, ok := s.Read()
				if !ok {
					continue
				}
				atomic.AddInt64(&benchSink, v.A)
				runtime.KeepAlive(v)
			}
		}
	})
}

func runMixedPointer(b *testing.B, writeEvery int) {
	p := zutil.NewPointer(unsafe.Pointer(&benchData{}))
	var ctr int64

	b.ReportAllocs()
	b.SetParallelism(4)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			if writeEvery > 0 && i%writeEvery == 0 {
				v := atomic.AddInt64(&ctr, 1)
				d := &benchData{A: v, B: v, C: v}
				p.Store(unsafe.Pointer(d))
			} else {
				d := (*benchData)(p.Load())
				atomic.AddInt64(&benchSink, d.A)
				runtime.KeepAlive(d)
			}
		}
	})
}

func BenchmarkSeqLockT_Read(b *testing.B) {
	s := NewSeqLock[*benchData]()
	s.Write(&benchData{A: 1, B: 2, C: 3})

	b.ReportAllocs()
	b.SetParallelism(4)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			d, _ := s.Read()
			atomic.AddInt64(&benchSink, d.A)
			runtime.KeepAlive(d)
		}
	})
}

func BenchmarkSeqLockT_Write(b *testing.B) {
	s := NewSeqLock[*benchData]()
	var x int64

	b.ReportAllocs()
	b.SetParallelism(4)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v := atomic.AddInt64(&x, 1)
			s.Write(&benchData{A: v, B: v, C: v})
		}
	})
}

func BenchmarkSeqLockT_Mixed10(b *testing.B) {
	s := NewSeqLock[*benchData]()
	s.Write(&benchData{})
	var ctr int64

	b.ReportAllocs()
	b.SetParallelism(4)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			if i%10 == 0 {
				v := atomic.AddInt64(&ctr, 1)
				s.Write(&benchData{A: v, B: v, C: v})
			} else {
				d, _ := s.Read()
				atomic.AddInt64(&benchSink, d.A)
				runtime.KeepAlive(d)
			}
		}
	})
}

// Single write, then many reads
func BenchmarkSeqLock_ReadOnce(b *testing.B) {
	s := NewSeqLock[*benchData]()
	s.Write(&benchData{A: 1, B: 2, C: 3})

	b.ReportAllocs()
	b.SetParallelism(4)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v, ok := s.Read()
			if !ok {
				continue
			}
			atomic.AddInt64(&benchSink, v.A)
			runtime.KeepAlive(v)
		}
	})
}

func BenchmarkPointer_ReadOnce(b *testing.B) {
	p := zutil.NewPointer(unsafe.Pointer(&benchData{A: 1, B: 2, C: 3}))

	b.ReportAllocs()
	b.SetParallelism(4)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			d := (*benchData)(p.Load())
			atomic.AddInt64(&benchSink, d.A)
			runtime.KeepAlive(d)
		}
	})
}
