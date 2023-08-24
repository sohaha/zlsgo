package zreflect

import (
	"errors"
	"reflect"
)

func ForEachMethod(valof reflect.Value, fn func(index int, method reflect.Method, value reflect.Value) error) error {
	numMethod := valof.NumMethod()
	if numMethod == 0 {
		return errors.New("method cannot be obtained")
	}

	tp := toRType(NewType(valof))
	var err error
	for i := 0; i < numMethod; i++ {
		err = fn(i, tp.Method(i), valof.Method(i))
		if err != nil {
			return err
		}
	}

	return nil
}

var (
	// SkipChild Field is returned when a struct field is skipped.
	SkipChild = errors.New("skip struct")
)

// ForEach For Each Struct field
func ForEach(typ reflect.Type, fn func(parent []string, index int, tag string, field reflect.StructField) error) (err error) {
	var forField func(typ reflect.Type, parent []string) error
	forField = func(typ reflect.Type, parent []string) error {
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			fieldTag, _ := GetStructTag(field)

			err = fn(parent, i, fieldTag, field)
			if err == SkipChild {
				continue
			}

			if err == nil && field.Type.Kind() == reflect.Struct {
				err = forField(field.Type, append(parent, fieldTag))
			}

			if err != nil {
				return err
			}
		}
		return nil
	}

	return forField(typ, []string{})
}

// ForEachValue For Each Struct field and value
func ForEachValue(val reflect.Value, fn func(parent []string, index int, tag string, field reflect.StructField, val reflect.Value) error) (err error) {
	if !Nonzero(val) {
		return errors.New("reflect.Value is zero")
	}

	typ := toRType(NewType(val))
	var forField func(val reflect.Value, typ reflect.Type, parent []string) error
	forField = func(val reflect.Value, typ reflect.Type, parent []string) error {
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			fieldValue := val.Field(i)
			fieldTag, _ := GetStructTag(field)
			if field.PkgPath != "" {
				continue
			}

			err = fn(parent, i, fieldTag, field, fieldValue)
			if err == SkipChild {
				continue
			}

			if err == nil && field.Type.Kind() == reflect.Struct {
				err = forField(fieldValue, field.Type, append(parent, fieldTag))
			}

			if err != nil {
				return err
			}
		}

		return nil
	}

	return forField(val, typ, []string{})
}
