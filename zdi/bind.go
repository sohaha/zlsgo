package zdi

import (
	"errors"
	"fmt"
	"reflect"
)

type (
	Construct interface {
		Apply(Pointer) error
		Resolve(...Pointer) error
	}
)

func (inj *injector) Resolve(v ...Pointer) error {
	for _, p := range v {
		r := reflect.ValueOf(p)
		rt := r.Type()
		v := r
		t := rt
		for v.Kind() == reflect.Ptr {
			v = v.Elem()
			t = rt.Elem()
		}

		if !v.IsValid() {
			var val reflect.Value
			var ok bool

			val, ok = inj.Get(t)
			if !ok {
				return errors.New("can't find injector for " + t.String())
			}

			if !v.IsValid() && r.IsValid() {
				if t.Kind() == reflect.Ptr {
					r.Elem().Set(val)
					continue
				}
			}

			return errors.New("invalid pointer")
		}

		if !v.CanSet() {
			return errors.New("cannot set value")
		}

		typ := v.Type()
		val, ok := inj.Get(typ)
		if !ok {
			return errors.New("can't find injector for " + t.String())
		}

		v.Set(val)
	}
	return nil
}

func (inj *injector) Apply(p Pointer) error {
	v := reflect.ValueOf(p)
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
