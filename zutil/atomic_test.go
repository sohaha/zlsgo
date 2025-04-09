package zutil_test

import (
	"math/rand"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zutil"
)

func TestBool(t *testing.T) {
	tt := zlsgo.NewTest(t)

	isOk := zutil.NewBool(false)
	tt.EqualTrue(!isOk.Load())

	isOk.Store(true)
	tt.EqualTrue(isOk.Load())

	isOk.Store(false)
	tt.EqualTrue(!isOk.Load())

	tt.EqualTrue(isOk.CAS(false, true))

	tt.EqualTrue(isOk.Load())

	// the current is true
	tt.EqualTrue(!isOk.CAS(false, true))

	tt.EqualTrue(isOk.CAS(true, true))

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			isOk.Toggle()
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestInt32(t *testing.T) {
	tt := zlsgo.NewTest(t)

	count := zutil.NewInt32(0)
	tt.EqualTrue(count.Load() == 0)

	var wg sync.WaitGroup
	l := rand.Intn(10000) + 10000

	count.Store(100)

	for i := 0; i < l; i++ {
		wg.Add(1)
		go func() {
			count.Add(1)
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			count.Sub(1)
			wg.Done()
		}()
	}
	wg.Wait()

	tt.Equal(int32(100), count.Load())
	count.Swap(200)
	tt.Equal(int32(200), count.Load())
	count.CAS(200, 300)
	tt.Equal(int32(300), count.Load())
}

func TestInt64(t *testing.T) {
	tt := zlsgo.NewTest(t)

	count := zutil.NewInt64(0)
	tt.EqualTrue(count.Load() == 0)

	var wg sync.WaitGroup
	l := rand.Intn(10000) + 10000

	count.Store(100)

	for i := 0; i < l; i++ {
		wg.Add(1)
		go func() {
			count.Add(1)
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			count.Sub(1)
			wg.Done()
		}()
	}
	wg.Wait()

	tt.Equal(int64(100), count.Load())
	count.Swap(200)
	tt.Equal(int64(200), count.Load())
	count.CAS(200, 300)
	tt.Equal(int64(300), count.Load())
}

func TestUint64(t *testing.T) {
	tt := zlsgo.NewTest(t)

	count := zutil.NewUint64(0)
	tt.EqualTrue(count.Load() == 0)

	var wg sync.WaitGroup
	l := rand.Intn(10000) + 10000

	count.Store(100)

	for i := 0; i < l; i++ {
		wg.Add(1)
		go func() {
			count.Add(1)
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			count.Sub(1)
			wg.Done()
		}()
	}
	wg.Wait()

	tt.Equal(uint64(100), count.Load())
	count.Swap(200)
	tt.Equal(uint64(200), count.Load())
	count.CAS(200, 300)
	tt.Equal(uint64(300), count.Load())
}

func TestUint32(t *testing.T) {
	tt := zlsgo.NewTest(t)

	count := zutil.NewUint32(0)
	tt.EqualTrue(count.Load() == 0)

	var wg sync.WaitGroup
	l := rand.Intn(10000) + 10000

	count.Store(100)

	for i := 0; i < l; i++ {
		wg.Add(1)
		go func() {
			count.Add(1)
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			count.Sub(1)
			wg.Done()
		}()
	}
	wg.Wait()

	tt.Equal(uint32(100), count.Load())
	count.Swap(200)
	tt.Equal(uint32(200), count.Load())
	count.CAS(200, 300)
	tt.Equal(uint32(300), count.Load())
}

func TestUintptr(t *testing.T) {
	tt := zlsgo.NewTest(t)

	v := &TestSt{
		Name: "test",
		I:    0,
	}

	f := zutil.NewPointer(unsafe.Pointer(v))

	tt.Equal(v, (*TestSt)(f.Load()))

	ii := zutil.NewInt64(0)
	var wg sync.WaitGroup
	for i := 0; i < 1010; i++ {
		wg.Add(1)
		go func() {
			ii.Add(1)
			a := (*TestSt)(f.Load())
			wg.Done()
			tt.Equal(v.Name, a.Name)
		}()
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			time.Sleep(time.Microsecond)
			ii.Add(1)
			a := &TestSt{
				Name: "test",
				I:    zstring.RandInt(0, 100),
			}
			f.Store(unsafe.Pointer(a))

			tt.EqualTrue(unsafe.Pointer(a) != unsafe.Pointer(v))
			wg.Done()
		}()
	}

	wg.Wait()
	tt.Log(ii.Load())
	tt.Log(f.String())

	tt.EqualTrue(!f.CAS(unsafe.Pointer(v), unsafe.Pointer(&TestSt{})))
}
