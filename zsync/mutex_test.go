package zsync

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestRBMutex(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var (
		wg    WaitGroup
		total = 100
		maps  = make(map[int]int)
		mu    = NewRBMutex()
	)

	for i := 0; i < total; i++ {
		maps[i] = i
	}

	for i := 0; i < total; i++ {
		ii := i
		wg.Go(func() {
			mu.Lock()
			maps[ii*2] = ii * 2
			mu.Unlock()
		})

		wg.Go(func() {
			t := mu.RLock()
			tt.Equal(ii, maps[ii])
			mu.RUnlock(t)
		})
	}

	wg.Wait()
}
