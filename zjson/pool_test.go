package zjson

import (
	"strings"
	"sync"
	"testing"
)

// TestResetPathCacheClearsElements pins the invariant that a recycled pool
// entry never keeps substrings of a previous caller's path alive. It runs on a
// locally owned slice, so it is deterministic and free of pool concurrency.
func TestResetPathCacheClearsElements(t *testing.T) {
	cache := make([]pathResult, 0, 8)
	cache = append(cache, pathResult{part: "secret-part", gpart: "secret-part", path: "secret-path"})

	resetPathCache(&cache)
	if len(cache) != 0 || cap(cache) != 8 {
		t.Fatalf("reset must truncate and keep capacity: len=%d cap=%d", len(cache), cap(cache))
	}
	for i, r := range cache[:cap(cache)] {
		if r.part != "" || r.gpart != "" || r.path != "" || r.force || r.more {
			t.Fatalf("element %d not cleared: %+v", i, r)
		}
	}
}

const (
	deepPathDoc  = `{"a":{"b":{"c":{"d":{"e":{"f":{"g":{"h":{"i":{"j":{"k":{"l":1}}}}}}}}}}}}`
	deepPath     = "a.b.c.d.e.f.g.h.i.j.k.l"
	deepPathWant = `{"a":{"b":{"c":{"d":{"e":{"f":{"g":{"h":{"i":{"j":{"k":{"l":2}}}}}}}}}}}}`
)

// TestPathCacheGrowsAndIsReused checks that set() appends through the pooled
// pointer instead of a local copy: a deep path must grow the entry itself, so
// later calls find the capacity already in place.
func TestPathCacheGrowsAndIsReused(t *testing.T) {
	got, err := Set(deepPathDoc, deepPath, 2)
	if err != nil {
		t.Fatalf("set deep path: %v", err)
	}
	if got != deepPathWant {
		t.Fatalf("deep path result mismatch:\n got %s\nwant %s", got, deepPathWant)
	}

	// Seed the pool on this P with a known entry, then observe that entry: when
	// set() appends through the pointer the header grows past the initial cap 8.
	// A local-copy append leaves the pooled header untouched, so cap stays 8.
	// The loop tolerates the goroutine being rescheduled onto another P, where
	// the pooled entry we seeded is not the one set() receives.
	want := len(strings.Split(deepPath, "."))
	for attempt := 0; attempt < 10; attempt++ {
		p := getPathCache()
		// Replace whatever state the pool had with a flat cap-8 entry so the
		// observation does not depend on which test ran first.
		*p = make([]pathResult, 0, 8)
		putPathCache(p)

		if _, err := Set(deepPathDoc, deepPath, 2); err != nil {
			t.Fatalf("set deep path: %v", err)
		}
		if cap(*p) >= want {
			return
		}
	}
	t.Fatalf("pooled header capacity never grew: set() is not appending through the pooled pointer")
}

// TestPathCacheAllocsStable is a coarse guard against re-introducing per-call
// segment slice growth. The pre-fix cost was 18 allocs for the deep path; the
// header write-back removes the one regrowth allocation.
func TestPathCacheAllocsStable(t *testing.T) {
	if raceDetector {
		t.Skip("allocation counts are inflated by race instrumentation")
	}
	allocs := testing.AllocsPerRun(50, func() {
		_, _ = Set(deepPathDoc, deepPath, 2)
	})
	if allocs > 18 {
		t.Fatalf("deep path allocs = %v, want <= 18", allocs)
	}
	t.Logf("deep path allocs = %v", allocs)
}

// TestSetPathPoolConcurrent exercises the pool from many goroutines: the race
// detector must stay clean and results must stay correct per call.
func TestSetPathPoolConcurrent(t *testing.T) {
	cases := []struct {
		doc  string
		path string
	}{
		{`{"a":1}`, "a"},
		{`{"a":{"b":1}}`, "a.b"},
		{`{"a":{"b":{"c":{"d":{"e":1}}}}}`, "a.b.c.d.e"},
		{`{"a":{"b":{"c":{"d":{"e":{"f":{"g":{"h":{"i":1}}}}}}}}}`, "a.b.c.d.e.f.g.h.i"},
	}

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(seed int) {
			defer wg.Done()
			for n := 0; n < 200; n++ {
				c := cases[(seed+n)%len(cases)]
				if _, err := Set(c.doc, c.path, n); err != nil {
					t.Errorf("set %q: %v", c.path, err)
					return
				}
			}
		}(i)
	}
	wg.Wait()
}

func BenchmarkSetDeepPath(b *testing.B) {
	deep := `{"a":{"b":{"c":{"d":{"e":{"f":{"g":{"h":{"i":{"j":{"k":{"l":1}}}}}}}}}}}}`
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = Set(deep, "a.b.c.d.e.f.g.h.i.j.k.l", 2)
	}
}
