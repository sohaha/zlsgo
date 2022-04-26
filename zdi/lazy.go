package zdi

import (
	"reflect"
)

func (inj *injector) Provide(provider interface{}, opt ...Option) (override []reflect.Type) {
	val := reflect.ValueOf(provider)
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
