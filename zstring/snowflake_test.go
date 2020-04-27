package zstring

import (
	"sync"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
)

func TestId(t *testing.T) {
	var g sync.WaitGroup
	var testSum = 10000
	var ids = struct {
		data map[int64]*interface{}
		sync.RWMutex
	}{
		data: make(map[int64]*interface{}, testSum),
	}

	tt := zlsgo.NewTest(t)
	w, err := NewIdWorker(0)
	tt.EqualNil(err)

	g.Add(testSum)
	for i := 0; i < testSum; i++ {
		go func() {
			id, err := w.Id()
			tt.EqualNil(err)
			ids.Lock()
			if _, ok := ids.data[id]; ok {
				t.Fatal("repeated")
			}
			ids.data[id] = new(interface{})
			ids.Unlock()
			g.Done()
		}()
	}
	g.Wait()

	w, err = NewIdWorker(2)
	tt.EqualNil(err)
	id, _ := w.Id()
	tim, ts, workerId, seq := ParseId(id)
	tt.EqualNil(err)
	t.Log(id, tim, ts, workerId, seq)
}

func TestIdWorker_timeReGen(t *testing.T) {
	tt := zlsgo.NewTest(t)
	w, err := NewIdWorker(0)
	tt.EqualNil(err)

	t.Log(w.Id())

	g := w.timeGen()
	now := time.Now()
	reG := w.timeReGen(g + 1)
	t.Log(g, reG)
	v := time.Now().Sub(now).Nanoseconds()

	g = w.timeGen()
	now = time.Now()
	reG = w.timeReGen(g)
	t.Log(g, reG)

	t.Log(v, time.Now().Sub(now).Nanoseconds())
}
