package zdi

import (
	"reflect"

	"github.com/sohaha/zlsgo/zreflect"
)

// Provide registers a provider function with the injector.
// A provider is a function that, when invoked, returns one or more values to be injected.
func (inj *injector) Provide(provider interface{}, opt ...Option) (override []reflect.Type) {
	val := zreflect.ValueOf(provider)
	t := val.Type()
	numout := t.NumOut()
	for i := 0; i < numout; i++ {
		out := t.Out(i)
		if _, ok := inj.values[out]; ok {
			override = append(override, out)
		}
		if _, ok := inj.providers[out]; ok {
			override = append(override, out)
		}
		inj.providers[out] = val
	}
	return
}
