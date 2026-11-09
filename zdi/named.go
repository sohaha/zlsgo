package zdi

import (
	"fmt"
	"reflect"
)

// GetNamed gets a value by name from the injector.
// If a provider exists for the name, it will be executed to obtain the value.
// Construction failures are not visible to the caller (return false); use
// GetNamedErr to observe them.
func GetNamed(inj Injector, name string, t reflect.Type) (reflect.Value, bool) {
	i, ok := inj.(*injector)
	if !ok {
		return reflect.Value{}, false
	}
	v, ok, _ := i.getNamed(name, t, nil)
	return v, ok
}

// GetNamedErr gets a value by name from the injector, returning error if any.
// This is the error-aware API for named dependencies.
func GetNamedErr(inj Injector, name string, t reflect.Type) (reflect.Value, bool, error) {
	i, ok := inj.(*injector)
	if !ok {
		return reflect.Value{}, false, fmt.Errorf("zdi: not a zdi injector")
	}
	return i.getNamed(name, t, nil)
}

// getNamed is the single resolution path for named dependencies, matching the
// type-based get semantics: failures are not written and the container can be
// retried.
func (inj *injector) getNamed(name string, t reflect.Type, stack depStack) (reflect.Value, bool, error) {
	token := inj.mu.RLock()
	val, hasVal := inj.named[name]
	_, hasProvider := inj.namedProviders[name]
	parent := inj.parent
	inj.mu.RUnlock(token)

	if hasVal && val.IsValid() {
		if t != nil && !val.Type().AssignableTo(t) {
			return reflect.Value{}, false, nil
		}
		return val, true, nil
	}

	if hasProvider {
		if stack.contains(namedKey(name)) {
			return reflect.Value{}, false, fmt.Errorf("zdi: circular dependency detected while resolving named %q", name)
		}
		next := stack.push(namedKey(name))

		key := "named:" + name
		call, leader, err := inj.beginBuild(key)
		if err != nil {
			return reflect.Value{}, false, err
		}
		if leader {
			buildErr := inj.runBuild(key, call, func() error {
				return inj.buildNamed(name, next)
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
		val, hasVal = inj.named[name]
		inj.mu.RUnlock(token)
		if hasVal && val.IsValid() {
			if t != nil && !val.Type().AssignableTo(t) {
				return reflect.Value{}, false, nil
			}
			return val, true, nil
		}
		return reflect.Value{}, false, nil
	}

	if parent != nil {
		if p, ok := parent.(*injector); ok {
			return p.getNamed(name, t, stack)
		}
	}

	return reflect.Value{}, false, nil
}

// buildNamed runs the provider registered under name and commits its result.
func (inj *injector) buildNamed(name string, stack depStack) error {
	token := inj.mu.RLock()
	if current, ok := inj.named[name]; ok && current.IsValid() {
		inj.mu.RUnlock(token)
		return nil
	}
	provider, ok := inj.namedProviders[name]
	inj.mu.RUnlock(token)
	if !ok || !provider.IsValid() {
		return nil
	}

	results, err := inj.invoke(provider.Interface(), stack)
	if err != nil {
		return err
	}

	errIdx := -1
	pt := provider.Type()
	for i := 0; i < pt.NumOut(); i++ {
		if pt.Out(i) == errorType {
			errIdx = i
		}
	}
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
		inj.named[name] = result
		break
	}
	inj.mu.Unlock()
	return nil
}
