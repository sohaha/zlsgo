package zdi_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zdi"
)

type (
	CycA struct{ B *CycB }
	CycB struct{ A *CycA }
)

func TestSemanticV2_Extended(t *testing.T) {
	tt := zlsgo.NewTest(t)

	t.Run("circular_dependency_returns_error", func(t *testing.T) {
		di := zdi.New()
		di.Provide(func(b *CycB) (*CycA, error) { return &CycA{B: b}, nil })
		di.Provide(func(a *CycA) (*CycB, error) { return &CycB{A: a}, nil })

		var a *CycA
		err := di.Resolve(&a)
		tt.Error(err)
		tt.EqualNil(a)
	})

	t.Run("named_provider_resolves", func(t *testing.T) {
		di := zdi.New()
		di.Provide(func() (string, error) {
			return "primary-db", nil
		}, zdi.Name("primary"))

		v, ok := zdi.GetNamed(di, "primary", reflect.TypeOf(""))
		tt.EqualTrue(ok)
		tt.Equal("primary-db", v.String())
	})

	t.Run("named_provider_failure_is_retriable", func(t *testing.T) {
		di := zdi.New()
		calls := 0
		di.Provide(func() (int, error) {
			calls++
			if calls == 1 {
				return 0, errors.New("named fail")
			}
			return 42, nil
		}, zdi.Name("answer"))

		v, ok, err := zdi.GetNamedErr(di, "answer", reflect.TypeOf(0))
		tt.Error(err)
		tt.False(ok)
		tt.False(v.IsValid())

		v, ok, err = zdi.GetNamedErr(di, "answer", reflect.TypeOf(0))
		tt.NoError(err)
		tt.EqualTrue(ok)
		tt.Equal(42, int(v.Int()))
		tt.Equal(2, calls)
	})

	t.Run("generic_Resolve", func(t *testing.T) {
		di := zdi.New()
		di.Map("hello")

		s, err := zdi.Resolve[string](di)
		tt.NoError(err)
		tt.Equal("hello", s)
	})

	t.Run("generic_Resolve_named", func(t *testing.T) {
		di := zdi.New()
		di.Map("named-value", zdi.Name("cfg"))

		s, err := zdi.ResolveNamed[string](di, "cfg")
		tt.NoError(err)
		tt.Equal("named-value", s)
	})

	t.Run("GetErr_surfaces_construction_error", func(t *testing.T) {
		di := zdi.New()
		di.Provide(func() (*RT, error) {
			return nil, errors.New("explicit failure")
		})

		_, ok, err := zdi.GetErr(di, reflect.TypeOf((*RT)(nil)))
		tt.Error(err)
		tt.False(ok)

		okDI := zdi.New()
		okDI.Map(&RT{OK: true})
		v, ok, err := zdi.GetErr(okDI, reflect.TypeOf((*RT)(nil)))
		tt.NoError(err)
		tt.EqualTrue(ok)
		tt.EqualTrue(v.Interface().(*RT).OK)
	})

	t.Run("provide_validation_panics", func(t *testing.T) {
		di := zdi.New()
		mustPanic := func(fn func()) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatal("expected panic but got none")
				}
			}()
			fn()
		}
		mustPanic(func() {
			di.Provide(func() error { return nil })
		})
		mustPanic(func() {
			di.Provide(func() (int, error, string) { return 0, nil, "" })
		})
		mustPanic(func() {
			di.Provide(func() {})
		})
		mustPanic(func() {
			di.Provide("not a func")
		})
	})

	// A provider must not be able to re-enter the same singleflight key: that
	// would block forever instead of reporting an error.
	t.Run("named_self_reference_returns_error_not_deadlock", func(t *testing.T) {
		di := zdi.New()
		typ := reflect.TypeOf((*CycA)(nil))
		di.Provide(func() (*CycA, error) {
			v, _, err := zdi.GetNamedErr(di, "self", typ)
			_ = v
			return &CycA{}, err
		}, zdi.Name("self"))

		done := make(chan error, 1)
		go func() {
			_, _, err := zdi.GetNamedErr(di, "self", typ)
			done <- err
		}()

		select {
		case err := <-done:
			if err == nil {
				t.Fatal("expected circular dependency error, got nil")
			}
		case <-time.After(3 * time.Second):
			t.Fatal("named self reference deadlocked instead of returning an error")
		}
	})
}
