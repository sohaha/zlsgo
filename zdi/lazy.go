package zdi

import (
	"fmt"
	"reflect"

	"github.com/sohaha/zlsgo/zreflect"
)

// Provide registers a provider function with the injector.
//
// Semantics:
//   - A constructor may return (T, error), (T1, T2, error), etc. The error must
//     be the last return value.
//   - error is construction control flow, never an injectable value.
//   - Construction failures are not cached; the next Get/Resolve retries.
//   - A function returning only error is not a provider; use InvokeWithErrorOnly
//     instead, Provide panics on it.
func (inj *injector) Provide(provider interface{}, opt ...Option) (override []reflect.Type) {
	val := zreflect.ValueOf(provider)
	if !val.IsValid() || val.Kind() != reflect.Func {
		panic(fmt.Sprintf("zdi: Provide expects a function, got %T", provider))
	}
	t := val.Type()
	numout := t.NumOut()
	if numout == 0 {
		panic("zdi: provider must return at least one value: " + t.String())
	}

	errIdx := -1
	for i := 0; i < numout; i++ {
		if t.Out(i) != errorType {
			continue
		}
		if i != numout-1 {
			panic("zdi: provider error return value must be the last one: " + t.String())
		}
		if errIdx != -1 {
			panic("zdi: provider may have only one error return value: " + t.String())
		}
		errIdx = i
	}
	if errIdx == 0 {
		panic("zdi: func() error is not a valid provider, use InvokeWithErrorOnly instead: " + t.String())
	}

	o := resolveOption(opt)
	inj.mu.Lock()
	defer inj.mu.Unlock()

	if o.name != "" {
		inj.namedProviders[o.name] = val
		return
	}

	for i := 0; i < numout; i++ {
		if i == errIdx {
			continue
		}
		out := t.Out(i)
		if _, ok := inj.values[out]; ok {
			override = append(override, out)
		}
		if _, ok := inj.providers[out]; ok {
			override = append(override, out)
		}
		inj.providers[out] = val
		inj.providerErrIdx[out] = errIdx
	}
	return
}
