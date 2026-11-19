package zjson

import (
	"sync"
	"testing"
)

// TestGetResConcurrentRead guards the read-only contract of *Res: numeric
// accessors must not mutate the Res, so a parsed value can be shared across
// goroutines. A regression that caches the parsed number on first read shows up
// here (and under -race, which CI runs).
func TestGetResConcurrentRead(t *testing.T) {
	const doc = `{"n":123,"t":[10,20,30],"s":"7"}`
	r := Get(doc, "n")
	arr := Get(doc, "t")

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				if got := r.Float64(); got != 123 {
					t.Errorf("Float64() = %v, want 123", got)
					return
				}
				if got := r.Int(); got != 123 {
					t.Errorf("Int() = %d, want 123", got)
					return
				}
				if got := r.String(); got != "123" {
					t.Errorf("String() = %q, want %q", got, "123")
					return
				}
				if got := arr.String(); got != "[10,20,30]" {
					t.Errorf("array String() = %q, want %q", got, "[10,20,30]")
					return
				}
			}
		}()
	}
	wg.Wait()
}
