package limiter

import (
	"fmt"
	"testing"
	"time"
)

func TestRecoveryPreservesActiveIndex(t *testing.T) {
	r := createRule(time.Minute, time.Second, 1, 1)
	keepKey := "keep"

	if err := r.add(keepKey); err != nil {
		t.Fatalf("add keep key failed: %v", err)
	}
	for i := 0; i < 9; i++ {
		if err := r.add(fmt.Sprintf("k%d", i)); err != nil {
			t.Fatalf("add key failed: %v", err)
		}
	}

	expired := time.Now().Add(-time.Minute).UnixNano()
	r.usedRecordsIndex.Range(func(k, v interface{}) bool {
		if k == keepKey {
			return true
		}
		idx := v.(int)
		if idx >= 0 && idx < len(r.records) {
			q := r.records[idx]
			if q.head != q.tail {
				q.slice[q.head] = expired
			}
		}
		return true
	})

	r.deleteExpiredOnce()
	if !r.needRecovery() {
		t.Fatalf("expected recovery to be needed")
	}
	r.recovery()

	remaining := r.remainingVisits(keepKey)
	if remaining != 0 {
		t.Fatalf("expected remaining 0, got %d", remaining)
	}

	idxAny, ok := r.usedRecordsIndex.Load(keepKey)
	if !ok {
		t.Fatalf("missing keep key after recovery")
	}
	idx := idxAny.(int)
	if idx < 0 || idx >= len(r.records) {
		t.Fatalf("index out of range after recovery: %d", idx)
	}
}
