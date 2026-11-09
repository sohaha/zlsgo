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
		if e, ok := v[i].Interface().(error); ok && e != nil {
			return e
		}
	}

	return nil
}

// Invoke calls the given function, injecting its dependencies.
// It resolves the function's arguments using the injector's known types.
// Returns the function's return values and an error if dependency resolution or function call fails.
func (inj *injector) Invoke(f interface{}) ([]reflect.Value, error) {
	return inj.invoke(f, nil)
}

// invoke is the internal implementation of Invoke. It carries the current
// resolution stack so that circular dependencies can be detected.
func (inj *injector) invoke(f interface{}, stack depStack) (values []reflect.Value, err error) {
	catch := zerror.TryCatch(func() error {
		t := zreflect.TypeOf(f)
		switch v := f.(type) {
		case PreInvoker:
			values, err = inj.fast(v, t, t.NumIn(), stack)
		default:
			values, err = inj.call(f, t, t.NumIn(), stack)
		}
		return nil
	})

	if catch != nil {
		err = catch
	}

	return
}

// call is an internal helper to invoke a regular function.
// It resolves dependencies for the function's arguments and calls it.
func (inj *injector) call(f interface{}, t reflect.Type, numIn int, stack depStack) ([]reflect.Value, error) {
	var in []reflect.Value
	if numIn > 0 {
		in = make([]reflect.Value, numIn)
		for i := 0; i < numIn; i++ {
			argType := t.In(i)
			val, ok, err := inj.get(argType, stack)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, fmt.Errorf("value not found for type %v", argType)
			}

			in[i] = val
		}
	}
	return zreflect.ValueOf(f).Call(in), nil
}

// Map maps a value to its own type or to a specified interface type within the injector.
// Options can be provided, e.g., WithInterface or Name, to specify the mapping key.
// Returns the type key if it overrides an existing mapping.
func (inj *injector) Map(val interface{}, opt ...Option) (override reflect.Type) {
	o := resolveOption(opt)
	inj.mu.Lock()
	defer inj.mu.Unlock()

	if o.name != "" {
		inj.named[o.name] = zreflect.ValueOf(val)
		return nil
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
	inj.mu.Lock()
	inj.values[typ] = val
	inj.mu.Unlock()
}

// Get retrieves a value of the given type from the injector.
//
// It is the low-level compat entry, with the contract (value, ok): missing or
// construction failure both return ok=false and never panic. To observe
// construction errors use Resolve / Apply / Invoke / GetErr.
func (inj *injector) Get(t reflect.Type) (reflect.Value, bool) {
	val, ok, _ := inj.get(t, nil)
	if !ok || !val.IsValid() {
		return reflect.Value{}, false
	}
	return val, true
}

// GetErr follows the same resolution path as Get but surfaces construction
// errors and circular dependencies instead of swallowing them. It is the
// error-aware variant aimed at framework authors.
func (inj *injector) GetErr(t reflect.Type) (reflect.Value, bool, error) {
	return inj.get(t, nil)
}

// GetErr is the package-level counterpart of the GetErr method, usable with any
// Injector implementation and symmetric with GetNamedErr.
func GetErr(inj Injector, t reflect.Type) (reflect.Value, bool, error) {
	if i, ok := inj.(*injector); ok {
		return i.get(t, nil)
	}
	v, ok := inj.Get(t)
	return v, ok, nil
}

// get is the single value-resolution path. A construction failure returns an
// error; the failed value is never written into the container and the provider
// is not removed, so the next call re-runs the constructor.
func (inj *injector) get(t reflect.Type, stack depStack) (reflect.Value, bool, error) {
	token := inj.mu.RLock()
	val, hasVal := inj.values[t]
	provider, hasProvider := inj.providers[t]
	parent := inj.parent
	inj.mu.RUnlock(token)

	if hasVal && val.IsValid() {
		return val, true, nil
	}

	if hasProvider {
		if stack.contains(typeKey(t)) {
			return reflect.Value{}, false, fmt.Errorf("zdi: circular dependency detected while resolving %v", t)
		}

		next := stack.push(typeKey(t))
		key := providerKey(provider, t)

		call, leader, err := inj.beginBuild(key)
		if err != nil {
			return reflect.Value{}, false, err
		}
		if leader {
			buildErr := inj.runBuild(key, call, func() error {
				return inj.buildProvider(t, next)
			})
			if buildErr != nil {
				return reflect.Value{}, false, buildErr
			}
		} else {
			<-call.done
			if call.err != nil {
				return reflect.Value{}, false, call.err
			}
		}

		token = inj.mu.RLock()
		val, hasVal = inj.values[t]
		inj.mu.RUnlock(token)
		if hasVal && val.IsValid() {
			return val, true, nil
		}
		return reflect.Value{}, false, nil
	}

	// The error interface does not participate in concrete-type fuzzy matching,
	// so an arbitrary error implementation in the container is never
	// misinterpreted as an error slot.
	if t.Kind() == reflect.Interface && t != errorType {
		token = inj.mu.RLock()
		for k, v := range inj.values {
			if k.Implements(t) {
				val = v
				break
			}
		}
		inj.mu.RUnlock(token)
		if val.IsValid() {
			return val, true, nil
		}
	}

	if parent != nil {
		if p, ok := parent.(*injector); ok {
			return p.get(t, stack)
		}
		v, ok := parent.Get(t)
		return v, ok, nil
	}

	return reflect.Value{}, false, nil
}

// buildProvider runs the provider registered for t and commits its results.
// Nothing is written unless every value is produced and the error is nil.
func (inj *injector) buildProvider(t reflect.Type, stack depStack) error {
	token := inj.mu.RLock()
	if current, ok := inj.values[t]; ok && current.IsValid() {
		inj.mu.RUnlock(token)
		return nil
	}
	provider, ok := inj.providers[t]
	errIdx := inj.providerErrIdx[t]
	inj.mu.RUnlock(token)
	if !ok || !provider.IsValid() {
		return nil
	}

	results, err := inj.invoke(provider.Interface(), stack)
	if err != nil {
		return err
	}

	// Fully scan the error before writing anything, so partial results never
	// become cached singletons.
	if errIdx >= 0 && errIdx < len(results) {
		if e, ok := results[errIdx].Interface().(error); ok && e != nil {
			return e
		}
	}

	inj.mu.Lock()
	for i, result := range results {
		if i == errIdx || !result.IsValid() {
			continue
		}
		resultType := result.Type()
		inj.values[resultType] = result
		delete(inj.providers, resultType)
		delete(inj.providerErrIdx, resultType)
	}
	inj.mu.Unlock()
	return nil
}

func providerKey(provider reflect.Value, t reflect.Type) string {
	if provider.IsValid() && provider.Kind() == reflect.Func {
		return fmt.Sprintf("%#x", provider.Pointer())
	}
	return t.PkgPath() + ":" + t.String()
}
