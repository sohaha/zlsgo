//go:build go1.18
// +build go1.18

package zarray_test

import (
	"sync"
	"testing"

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

	m.Delele(2)

	tt.Equal(2, int(m.Len()))

	item, ok = m.Get(2)
	tt.EqualTrue(!ok)
	tt.Equal("", item)

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

func TestHashMapProvideGet(t *testing.T) {
	tt := zlsgo.NewTest(t)
	m := zarray.NewHashMap[int, int]()

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			v, ok := m.ProvideGet(1, func(i int) (int, bool) {
				return 99, true
			})
			tt.EqualTrue(ok)
			tt.Equal(99, v)
			wg.Done()
		}(i)
	}

	wg.Wait()

	tt.Log(m.Get(1))
}
