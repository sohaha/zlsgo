package zdi

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/sohaha/zlsgo/zreflect"
)

// Resolve resolves dependencies for the given pointers.
// It iterates through each pointer, determines its underlying type,
// and injects the corresponding value from the injector.
// Returns an error if a dependency cannot be found or if a value cannot be set.
func (inj *injector) Resolve(v ...Pointer) error {
	for _, p := range v {
		ptrValue := zreflect.ValueOf(p)
		if ptrValue.Kind() != reflect.Ptr {
			return errors.New("cannot resolve non-pointer value, argument must be a pointer to the target variable")
		}

		elemToSet := ptrValue.Elem()

		if !elemToSet.IsValid() {
			targetTypeForNil := ptrValue.Type().Elem()
			_, ok := inj.Get(targetTypeForNil)
			if !ok {
				return errors.New("can't find injector for " + targetTypeForNil.String())
			}
			return errors.New("cannot set to an uninitialized (nil) pointer: " + targetTypeForNil.String())
		}

		if !elemToSet.CanSet() {
			return errors.New("cannot set underlying value of pointer: " + elemToSet.String())
		}

		targetType := elemToSet.Type()

		resolvedVal, ok := inj.Get(targetType)
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

// Apply injects dependencies into the fields of a struct or sets a pointer value.
// If the provided pointer is a struct, it iterates through its fields.
// For fields tagged with `di`, it resolves and injects the corresponding dependency.
// If the pointer is not a struct, it attempts to resolve and set the value directly.
// Returns an error if a dependency cannot be found for a tagged field or if a value cannot be set.
func (inj *injector) Apply(p Pointer) error {
	v := zreflect.ValueOf(p)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if !v.CanSet() {
		return errors.New("cannot set value")
	}

	typ := v.Type()
	val, ok := inj.Get(typ)
	if ok {
		v.Set(val)
		return nil
	}

	if v.Kind() != reflect.Struct {
		val, ok := inj.Get(typ)
		if !ok {
			return nil
		}
		v.Set(val)
		return nil
	}

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		structField := typ.Field(i)
		_, ok := structField.Tag.Lookup("di")
		if f.CanSet() && ok {
			ft := f.Type()
			v, ok := inj.Get(ft)
			if !ok {
				return fmt.Errorf("value not found for type %v", ft)
			}
			f.Set(v)
		}
	}
	return nil
}
