//go:build go1.18
// +build go1.18

package zarray_test

import (
	"sync"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zarray"
)

func TestHashMap(t *testing.T) {
	tt := zlsgo.NewTest(t)
	m := zarray.NewHashMap[int, string]()

	m.Set(1, "name")
	m.Set(2, "name2")
	m.Set(9, "name9")
	tt.Equal(3, int(m.Len()))

	item, ok := m.Get(1)
	tt.EqualTrue(ok)
	tt.Equal("name", item)

	item, ok = m.Get(2)
	tt.EqualTrue(ok)
	tt.Equal("name2", item)

	item, ok = m.Get(3)
	tt.EqualTrue(!ok)
	tt.Equal("", item)

	m.Delete(2, 9)

	tt.Equal(1, int(m.Len()))

	item, ok = m.Get(2)
	tt.EqualTrue(!ok)

	m.ForEach(func(key int, value string) bool {
		t.Log(key, value)
		return true
	})
}

func TestHashMapOverwrite(t *testing.T) {
	tt := zlsgo.NewTest(t)
	m := zarray.NewHashMap[int, string]()
	key := 1
	name := "luffy"
	name2 := "ace"

	m.Set(key, name)
	m.Set(key, name2)
	tt.Equal(1, int(m.Len()))

	item, ok := m.Get(key)
	tt.EqualTrue(ok)
	tt.EqualTrue(name != item)
	tt.Equal(name2, item)
}

func TestHashMapSwap(t *testing.T) {
	tt := zlsgo.NewTest(t)
	m := zarray.NewHashMap[int, int]()
	m.Set(1, 100)

	oldValue, swapped := m.Swap(1, 200)
	t.Log(oldValue, swapped)
	tt.EqualTrue(swapped)
	tt.Equal(100, oldValue)

	oldValue, swapped = m.Swap(1, 200)
	t.Log(oldValue, swapped)

	tt.EqualTrue(!m.CAS(1, 100, 200))
	tt.EqualTrue(m.CAS(1, 200, 100))
	tt.EqualTrue(m.CAS(1, 100, 200))
	tt.EqualTrue(!m.CAS(1, 100, 200))
}

func TestHashMapProvideGet(t *testing.T) {
	tt := zlsgo.NewTest(t)
	m := zarray.NewHashMap[int, int]()

	{
		i := 0
		one, ok := m.ProvideGet(0, func() (int, bool) {
			i++
			return 110, false
		})
		t.Log(one, ok)

		one, ok = m.ProvideGet(0, func() (int, bool) {
			i++
			return 119, true
		})
		t.Log(one, ok)

		one, ok = m.ProvideGet(0, func() (int, bool) {
			i++
			return 119, true
		})
		t.Log(one, ok)
		tt.Equal(2, i)
	}

	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			v, ok := m.ProvideGet(1, func() (int, bool) {
				t.Log("set", 99)
				time.Sleep(time.Millisecond * 100)
				return 99, true
			})
			tt.EqualTrue(ok)
			tt.Equal(99, v)
			wg.Done()
		}(i)
	}

	wg.Wait()

	v, ok := m.Get(1)
	tt.EqualTrue(ok)
	tt.Equal(99, v)

	m.Delete(1)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			v, ok := m.ProvideGet(1, func() (int, bool) {
				time.Sleep(time.Millisecond * 100)
				t.Log("new set", 100)
				return 100, true
			})
			tt.EqualTrue(ok)
			tt.Equal(100, v)
			wg.Done()
		}(i)
	}

	wg.Wait()

	v, ok = m.Get(1)
	tt.EqualTrue(ok)
	tt.Equal(100, v)
}
