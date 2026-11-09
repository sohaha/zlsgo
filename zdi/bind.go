package zdi

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/sohaha/zlsgo/zreflect"
)

// Resolve resolves dependencies for the given pointers.
//
// Construction failures, missing dependencies, and circular dependencies are
// returned as error and never panic or inject partial values. Failures never
// pollute the container; the container can be fixed and Resolve can be retried.
func (inj *injector) Resolve(v ...Pointer) error {
	for _, p := range v {
		ptrValue := zreflect.ValueOf(p)
		if ptrValue.Kind() != reflect.Ptr {
			return errors.New("cannot resolve non-pointer value, argument must be a pointer to the target variable")
		}

		elemToSet := ptrValue.Elem()

		if !elemToSet.IsValid() {
			targetTypeForNil := ptrValue.Type().Elem()
			_, ok, err := inj.get(targetTypeForNil, nil)
			if err != nil {
				return err
			}
			if !ok {
				return errors.New("can't find injector for " + targetTypeForNil.String())
			}
			return errors.New("cannot set to an uninitialized (nil) pointer: " + targetTypeForNil.String())
		}

		if !elemToSet.CanSet() {
			return errors.New("cannot set underlying value of pointer: " + elemToSet.String())
		}

		targetType := elemToSet.Type()

		resolvedVal, ok, err := inj.get(targetType, nil)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("can't find injector for " + targetType.String())
		}

		if !resolvedVal.Type().AssignableTo(targetType) {
			return fmt.Errorf("resolved value of type %s is not assignable to target type %s", resolvedVal.Type().String(), targetType.String())
		}

		elemToSet.Set(resolvedVal)
	}
	return nil
}

// Apply injects dependencies into a struct or a pointer value.
// If the pointer is a struct, it iterates over fields. Fields tagged with `di`
// are resolved and injected. If the pointer is not a struct, it attempts to
// resolve and set the value directly.
// Returns an error if a dependency cannot be found for a tagged field or if a
// value cannot be set.
func (inj *injector) Apply(p Pointer) error {
	v := zreflect.ValueOf(p)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if !v.IsValid() || !v.CanSet() {
		return errors.New("cannot set value")
	}

	typ := v.Type()
	val, ok, err := inj.get(typ, nil)
	if err != nil {
		return err
	}
	if ok {
		v.Set(val)
		return nil
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		structField := typ.Field(i)
		_, hasTag := structField.Tag.Lookup("di")
		if !f.CanSet() || !hasTag {
			continue
		}
		ft := f.Type()
		val, ok, err := inj.get(ft, nil)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("value not found for type %v", ft)
		}
		f.Set(val)
	}
	return nil
}
