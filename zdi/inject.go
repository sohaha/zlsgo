package zdi

import (
	"fmt"
	"reflect"

	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zreflect"
)

// InvokeWithErrorOnly calls the given function and returns only an error, if any.
// It's a convenience wrapper around Invoke for functions that primarily signal success/failure via an error.
func (inj *injector) InvokeWithErrorOnly(f interface{}) (err error) {
	v, err := inj.Invoke(f)
	if err != nil {
		return err
	}

	if len(v) == 0 {
		return nil
	}

	for i := range v {
		if err, ok := v[i].Interface().(error); ok {
			return err
		}
	}

	return nil
}

// Invoke calls the given function, injecting its dependencies.
// It resolves the function's arguments using the injector's known types.
// Returns the function's return values and an error if dependency resolution or function call fails.
func (inj *injector) Invoke(f interface{}) (values []reflect.Value, err error) {
	catch := zerror.TryCatch(func() error {
		t := zreflect.TypeOf(f)
		switch v := f.(type) {
		case PreInvoker:
			values, err = inj.fast(v, t, t.NumIn())
		default:
			values, err = inj.call(f, t, t.NumIn())
		}
		return nil
	})

	if catch != nil {
		err = catch
	}

	return
}

// call is an internal helper to invoke a regular function.
// It resolves dependencies for the function's arguments and calls the function.
func (inj *injector) call(f interface{}, t reflect.Type, numIn int) ([]reflect.Value, error) {
	var in []reflect.Value
	if numIn > 0 {
		in = make([]reflect.Value, numIn)
		var argType reflect.Type
		for i := 0; i < numIn; i++ {
			argType = t.In(i)
			val, ok := inj.Get(argType)
			if !ok {
				return nil, fmt.Errorf("value not found for type %v", argType)
			}

			in[i] = val
		}
	}
	return zreflect.ValueOf(f).Call(in), nil
}

// Map maps a value to its own type or to a specified interface type within the injector.
// Options can be provided, e.g., WithInterface, to specify the mapping key.
// Returns the type key if it overrides an existing mapping.
func (inj *injector) Map(val interface{}, opt ...Option) (override reflect.Type) {
	o := mapOption{}
	for _, opt := range opt {
		opt(&o)
	}
	if o.key == nil {
		o.key = reflect.TypeOf(val)
	}
	if _, ok := inj.values[o.key]; ok {
		override = o.key
	}

	inj.values[o.key] = zreflect.ValueOf(val)
	return
}

// Maps maps multiple values into the injector.
// It's a convenience function for calling Map for each provided value.
// Returns a slice of types for which existing mappings were overridden.
func (inj *injector) Maps(values ...interface{}) (override []reflect.Type) {
	for _, val := range values {
		o := inj.Map(val)
		if o != nil {
			override = append(override, o)
		}
	}
	return
}

// Set maps a reflect.Value to a reflect.Type in the injector.
// This is a lower-level way to directly insert values into the injector's store.
func (inj *injector) Set(typ reflect.Type, val reflect.Value) {
	inj.values[typ] = val
}

// Get retrieves a value of the given type from the injector.
// It first checks for directly mapped values. If not found, it checks providers.
// If still not found and the type is an interface, it searches for compatible concrete types.
// Finally, it consults the parent injector, if one exists.
// Returns the reflect.Value and a boolean indicating if the value was found.
func (inj *injector) Get(t reflect.Type) (reflect.Value, bool) {
	val := inj.values[t]
	if val.IsValid() {
		return val, true
	}

	if provider, ok := inj.providers[t]; ok {
		results, err := inj.Invoke(provider.Interface())
		if err != nil {
			panic(err)
		}
		for _, result := range results {
			resultType := result.Type()
			inj.values[resultType] = result
			delete(inj.providers, resultType)
			if resultType == t {
				val = result
			}
		}

		if val.IsValid() {
			return val, true
		}
	}

	if t.Kind() == reflect.Interface {
		for k, v := range inj.values {
			if k.Implements(t) {
				val = v
				break
			}
		}
	}

	if val.IsValid() {
		return val, true
	}

	var ok bool
	if inj.parent != nil {
		val, ok = inj.parent.Get(t)
	}

	return val, ok
}
