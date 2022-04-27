package zpool

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zerror"
)

type preTest func(i int)

func (p preTest) Invoke(i []interface{}) ([]reflect.Value, error) {
	p(i[0].(int))
	return nil, nil
}

var _ zdi.PreInvoker = (*preTest)(nil)

func TestPoolInvoke(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var wg sync.WaitGroup
	testErr := errors.New("test")
	count := 10
	p := New(count)

	p.Injector().Map(1)
	p.Injector().Map(time.Now())

	p.PanicFunc(func(err error) {
		code, b := zerror.UnwrapCode(err)
		if b {
			t.Log("is zerror", code, err)
		} else {
			t.Log("is not zerror", err)
			if err != testErr {
				wg.Done()
			}
		}
	})

	wg.Add(1)
	err := p.Do(func() error {
		defer wg.Done()
		return testErr
	})
	tt.NoError(err)

	wg.Add(1)
	err = p.Do(preTest(func(i int) {
		defer wg.Done()
	}))
	tt.NoError(err)

	for i := 0; i <= 100; i++ {
		wg.Add(1)
		index := i
		err = p.Do(func(now time.Time) error {
			defer wg.Done()
			if index%20 == 0 {
				return zerror.New(zerror.ErrCode(index), now.String())
			}
			return nil
		})
		tt.NoError(err)
	}

	wg.Wait()
	time.Sleep(time.Second / 2)
}

func BenchmarkInjectorNo(b *testing.B) {
	var wg sync.WaitGroup
	count := 10000
	p := New(count)
	_ = p.PreInit()
	ii := 1
	wg.Add(b.N)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		err := p.Do(func() {
			_ = ii
			wg.Done()
		})
		if err != nil {
			b.Error(err)
		}
	}
	wg.Wait()
}

func BenchmarkInjectorPre(b *testing.B) {
	var wg sync.WaitGroup
	count := 10000
	p := New(count)
	p.Injector().Map(1)
	_ = p.PreInit()
	wg.Add(b.N)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		err := p.Do(preTest(func(ii int) {
			_ = ii
			wg.Done()
		}))
		if err != nil {
			b.Error(err)
		}
	}
	wg.Wait()
}

func BenchmarkInjector(b *testing.B) {
	var wg sync.WaitGroup
	count := 10000
	p := New(count)
	_ = p.PreInit()
	p.Injector().Map(1)
	wg.Add(b.N)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		err := p.Do(func(ii int) {
			_ = ii
			wg.Done()
		})
		if err != nil {
			b.Error(err)
		}
	}
	wg.Wait()
}
