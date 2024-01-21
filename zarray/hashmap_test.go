//go:build go1.18
// +build go1.18

package zarray_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zarray"
	"github.com/sohaha/zlsgo/zsync"
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

	m.Delete(2)
	m.Delete(9)

	tt.Equal(1, int(m.Len()))

	_, ok = m.Get(2)
	tt.EqualTrue(!ok)

	m.Set(2, "reset name2")
	m.ForEach(func(key int, value string) bool {
		t.Log("ForEach:", key, value)
		return true
	})

	j, err := json.Marshal(m)
	tt.NoError(err)
	t.Log(string(j))

	j = []byte(`{"2":"hobby","1":"new name","8":"886"}`)
	err = json.Unmarshal(j, &m)
	tt.NoError(err)
	mlen := m.Len()

	v2, ok := m.Get(2)
	tt.EqualTrue(ok)
	tt.Equal("hobby", v2)

	v1, ok := m.GetAndDelete(1)
	tt.EqualTrue(ok)
	tt.Equal("new name", v1)
	tt.Equal(mlen-1, m.Len())

	m.ForEach(func(key int, value string) bool {
		t.Log("n:", key, value)
		return true
	})

	t.Log(m.Keys())
	m.Clear()
	t.Log(m.Keys())
	m.Set(9, "yes")
	t.Log(m.Keys())
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
		one, ok, computed := m.ProvideGet(0, func() (int, bool) {
			t.Log("ProvideGet set", 110)
			i++
			return 110, false
		})
		t.Log(one, ok)
		tt.EqualTrue(!computed)

		one, ok, computed = m.ProvideGet(0, func() (int, bool) {
			t.Log("ProvideGet set", 119)
			i++
			return 119, true
		})
		t.Log(one, ok)
		tt.EqualTrue(computed)

		one, ok, computed = m.ProvideGet(0, func() (int, bool) {
			i++
			return 120, true
		})
		tt.EqualTrue(!computed)

		tt.EqualTrue(ok)
		tt.Equal(119, one)
		tt.Equal(2, i)
	}

	var wg zsync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Go(func() {
			v, ok, _ := m.ProvideGet(1, func() (int, bool) {
				t.Log("set", 99)
				time.Sleep(time.Millisecond * 100)
				return 99, true
			})
			tt.EqualTrue(ok)
			tt.Equal(99, v)
		})
	}

	_ = wg.Wait()

	v, ok := m.Get(1)
	tt.EqualTrue(ok)
	tt.Equal(99, v)

	m.Delete(1)

	for i := 0; i < 10; i++ {
		wg.Go(func() {
			v, ok, _ := m.ProvideGet(1, func() (int, bool) {
				time.Sleep(time.Millisecond * 100)
				t.Log("new set", 100)
				return 100, true
			})
			tt.EqualTrue(ok)
			tt.Equal(100, v)
		})
	}

	_ = wg.Wait()

	v, ok = m.Get(1)
	tt.EqualTrue(ok)
	tt.Equal(100, v)
}
