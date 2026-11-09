package zdi_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zdi"
)

// Semantics 2.0 locks three non-negotiable rules:
//  1. error is construction control flow, not an injectable value
//  2. construction failures must be visible (Resolve/Apply/Invoke return error)
//     and never silently swallowed
//  3. failures are never cached; the constructor re-runs after recovery and
//     panics become errors instead of escaping
type (
	RT  struct{ OK bool }
	Rep struct{}
	I18 struct{}
)

var (
	canFail = true
	rtCalls int
)

func newFail() (*RT, error) {
	rtCalls++
	if canFail {
		return nil, errors.New("boom: init failed")
	}
	return &RT{OK: true}, nil
}

func resetRT() { canFail, rtCalls = true, 0 }

func TestSemanticV2(t *testing.T) {
	tt := zlsgo.NewTest(t)

	t.Run("failure_is_not_silent_nil", func(t *testing.T) {
		resetRT()
		di := zdi.New()
		di.Provide(newFail)

		var rt *RT
		err := di.Resolve(&rt)
		tt.Error(err) // red baseline: Resolve returned nil
		tt.EqualNil(rt)

		_, ok := di.Get(reflect.TypeOf((*RT)(nil)))
		tt.False(ok) // red baseline: failure was cached as a nil singleton
	})

	t.Run("failure_not_cached_retryable", func(t *testing.T) {
		resetRT()
		di := zdi.New()
		di.Provide(newFail)

		var rt *RT
		_ = di.Resolve(&rt) // first call fails
		canFail = false
		_ = di.Resolve(&rt) // second call must re-run, not hit the error cache
		tt.NotNil(rt)       // red baseline: hit the cached nil
		tt.EqualTrue(rtCalls == 2) // red baseline: only constructed once
	})

	t.Run("multiple_error_providers", func(t *testing.T) {
		resetRT()
		di := zdi.New()
		di.Provide(func() (*Rep, error) { return &Rep{}, nil })
		di.Provide(func() (*I18, error) { return nil, errors.New("i18n: 建表失败") })
		di.Provide(newFail)

		var rep *Rep
		tt.NoError(di.Resolve(&rep)) // each provider's error applies independently
		tt.NotNil(rep)

		var i18 *I18
		tt.Error(di.Resolve(&i18)) // red baseline: err=nil
		tt.EqualNil(i18)

		var rt *RT
		tt.Error(di.Resolve(&rt)) // red baseline: err=nil
	})

	t.Run("panic_provider_returns_error", func(t *testing.T) {
		resetRT()
		di := zdi.New()
		di.Provide(func() *RT {
			if canFail {
				panic(errors.New("boom: init failed"))
			}
			return &RT{OK: true}
		})

		var rt *RT
		panicked := false
		func() {
			defer func() { if r := recover(); r != nil { panicked = true } }()
			_ = di.Resolve(&rt)
		}()
			tt.False(panicked) // red baseline: constructor panic escaped Resolve
			tt.EqualNil(rt)

			canFail = false // panic is not cached; recovery should succeed
			panicked = false
		func() {
			defer func() { if r := recover(); r != nil { panicked = true } }()
			_ = di.Resolve(&rt)
		}()
		tt.False(panicked)
		tt.NotNil(rt)
	})

	t.Run("multi_output_with_error_overrides_value", func(t *testing.T) {
		di := zdi.New()
		di.Provide(func() (*RT, error) {
			return &RT{OK: true}, errors.New("partial fail")
		})
		var rt *RT
			tt.Error(di.Resolve(&rt)) // red baseline: err=nil
			tt.EqualNil(rt)           // red baseline: value was injected, error ignored
	})

	t.Run("no_caching_of_failure", func(t *testing.T) {
		resetRT()
		di := zdi.New()
		di.Provide(newFail)

		var rt *RT
		_ = di.Resolve(&rt)
		_, ok := di.Get(reflect.TypeOf((*RT)(nil)))
			tt.False(ok) // red baseline: failure was cached
	})
}
