package zvalid

import (
	"sync"
	"testing"
)

// TestConcurrentRuleAppendAfterVerifiAny ensures that concurrent calls to VerifiAny
// do not race on the shared rule queue. The fix uses a full slice expression to
// converge the capacity, ensuring each call's queue lands in a private underlying array.
func TestConcurrentRuleAppendAfterVerifiAny(t *testing.T) {
	base := New().SetAlias("num")

	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			engine := base.VerifiAny(123)
			_, _ = engine.IsNumber().Int()
		}()
	}
	wg.Wait()
}
