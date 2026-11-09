package zdi

import (
	"fmt"
	"reflect"
)

// Resolve is a generic convenience function to resolve a dependency of type T.
// It automatically handles pointer dereferencing and returns the value and error.
func Resolve[T any](inj Injector) (T, error) {
	var out T
	if err := inj.Resolve(&out); err != nil {
		return out, err
	}
	return out, nil
}

// ResolveNamed is a generic convenience function to resolve a named dependency of type T.
// It looks up the name and resolves the dependency.
func ResolveNamed[T any](inj Injector, name string) (T, error) {
	var out T
	typ := reflect.TypeOf(&out).Elem()
	v, ok, err := GetNamedErr(inj, name, typ)
	if err != nil {
		return out, err
	}
	if !ok {
		return out, fmt.Errorf("zdi: can't find named injector %q", name)
	}
	if !v.Type().AssignableTo(typ) {
		return out, fmt.Errorf("zdi: named injector %q is %s, not %s", name, v.Type().String(), typ.String())
	}
	reflect.ValueOf(&out).Elem().Set(v)
	return out, nil
}