package zstring

import (
	"os"
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
	w, err := NewIDWorker(0)
	tt.EqualNil(err)

	g.Add(testSum)
	for i := 0; i < testSum; i++ {
		go func(t *testing.T) {
			id, err := w.ID()
			tt.EqualNil(err)
			ids.Lock()
			if _, ok := ids.data[id]; ok {
				t.Error("repeated")
				os.Exit(1)
			}
			ids.data[id] = new(interface{})
			ids.Unlock()
			g.Done()
		}(t)
	}
	g.Wait()

	w, err = NewIDWorker(2)
	tt.EqualNil(err)
	id, _ := w.ID()
	tim, ts, workerId, seq := ParseID(id)
	tt.EqualNil(err)
	t.Log(id, tim, ts, workerId, seq)
}

func TestIdWorker_timeReGen(t *testing.T) {
	tt := zlsgo.NewTest(t)
	w, err := NewIDWorker(0)
	tt.EqualNil(err)

	t.Log(w.ID())

	g := w.timeGen()
	now := time.Now()
	reG := w.timeReGen(g + 1)
	t.Log(g, reG)
	v := time.Since(now).Nanoseconds()

	g = w.timeGen()
	now = time.Now()
	reG = w.timeReGen(g)
	t.Log(g, reG)

	t.Log(v, time.Since(now).Nanoseconds())
}
