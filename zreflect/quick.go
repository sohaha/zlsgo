package zreflect

import (
	"errors"
	"reflect"
)

// ForEachMethod iterates through all methods of a given value and calls the provided function for each method.
// This simplifies the process of examining and working with methods via reflection.
//
// valof is the reflect.Value whose methods will be iterated.
// fn is a callback function that receives the method index, method information, and method value.
//
// It returns any error returned by the callback function, or nil if all iterations complete successfully.
func ForEachMethod(valof reflect.Value, fn func(index int, method reflect.Method, value reflect.Value) error) error {
	numMethod := valof.NumMethod()
	if numMethod == 0 {
		return nil
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

// SkipChild is a special error value that can be returned from ForEach and ForEachValue callbacks
// to indicate that the current struct field's children should be skipped during iteration.
var SkipChild = errors.New("skip struct")

// ForEach iterates through all fields of a struct type recursively and calls the provided function for each field.
// This provides a convenient way to examine and process struct fields via reflection.
//
// typ is the reflect.Type of the struct to iterate through.
// fn is a callback function that receives the parent path, field index, tag value, and field information.
//
// It returns any error returned by the callback function, or nil if all iterations complete successfully.
func ForEach(typ reflect.Type, fn func(parent []string, index int, tag string, field reflect.StructField) error) (err error) {
	var forField func(typ reflect.Type, parent []string) error
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
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

// ForEachValue iterates through all fields of a struct value recursively and calls the provided function for each field.
// Similar to ForEach, but also provides access to the field's value.
//
// val is the reflect.Value of the struct to iterate through.
// fn is a callback function that receives the parent path, field index, tag value, field information, and field value.
//
// It returns any error returned by the callback function, or nil if all iterations complete successfully.
func ForEachValue(val reflect.Value, fn func(parent []string, index int, tag string, field reflect.StructField, val reflect.Value) error) (err error) {
	if !Nonzero(val) {
		return errors.New("reflect.Value is zero")
	}

	val = reflect.Indirect(val)
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
