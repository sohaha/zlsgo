package zutil

import (
	"errors"
	"reflect"
	"strconv"
)

func SetValue(vTypeOf reflect.Kind, vValueOf reflect.Value, value interface{}) (err error) {
	typeErr := errors.New(vTypeOf.String() + " type assignment is not supported")
	vString := ""
	v, ok := value.(string)
	if ok {
		vString = v
	}
	switch vTypeOf {
	case reflect.String:
		vValueOf.SetString(vString)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(vString, 10, 64)
		if err != nil {
			err = errors.New("must be an integer")
		} else if vValueOf.OverflowInt(v) {
			err = typeErr
		}
		vValueOf.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(vString, 10, 64)
		if err != nil {
			err = errors.New("must be an unsigned integer")
		} else if vValueOf.OverflowUint(v) {
			err = typeErr
		}
		vValueOf.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(vString, 10)
		if err != nil {
			err = errors.New("must be decimal")
		} else if vValueOf.OverflowFloat(v) {
			err = typeErr
		}
	case reflect.Bool:
		v, err := strconv.ParseBool(vString)
		if err != nil {
			err = errors.New("must be boolean")
		}
		vValueOf.SetBool(v)
	case reflect.Slice:
		if value != nil {
			vValueOf.Set(reflect.ValueOf(value))
		} else {
			err = errors.New("must be slice")
		}
	case reflect.Struct:
		err = setStruct(vValueOf, value)
	default:
		err = typeErr
	}

	return err
}

// setStruct todo unfinished
func setStruct(v reflect.Value, value interface{}) (err error) {
	valueTypeof := reflect.TypeOf(value)
	kind := valueTypeof.Kind()
	if kind != reflect.Map {
		err = errors.New("must be map[]")
		return
	}

	if values, ok := value.(map[string]string); ok {
		ReflectForNumField(v, func(fieldTag string, kind reflect.Kind, field reflect.Value) bool {
			if v, ok := values[fieldTag]; ok {
				err = SetValue(kind, field, v)
			}
			return err == nil
		})
	} else {
		err = errors.New("not supported")
	}

	return
}

func ReflectForNumField(v reflect.Value, fn func(fieldTag string, kind reflect.Kind, field reflect.Value) bool, tag ...string) {
	tagKey := "z"
	if len(tag) > 0 {
		tagKey = tag[0]
	}
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		tfield := v.Type().Field(i)
		fieldTag := tfield.Tag.Get(tagKey)
		if fieldTag == "-" || !field.CanSet() {
			continue
		}
		fieldName := tfield.Name
		fieldType := field.Type()
		kind := fieldType.Kind()
		if fieldTag == "" {
			fieldTag = fieldName
		}
		if !fn(fieldTag, kind, field) {
			break
		}
	}
}

func Nonzero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return v.Float() != 0
	case reflect.Complex64, reflect.Complex128:
		return v.Complex() != complex(0, 0)
	case reflect.String:
		return v.String() != ""
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if Nonzero(GetField(v, i)) {
				return true
			}
		}
		return false
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if Nonzero(v.Index(i)) {
				return true
			}
		}
		return false
	case reflect.Map, reflect.Interface, reflect.Slice, reflect.Ptr, reflect.Chan, reflect.Func:
		return !v.IsNil()
	case reflect.UnsafePointer:
		return v.Pointer() != 0
	}
	return true
}

func GetField(v reflect.Value, i int) reflect.Value {
	val := v.Field(i)
	if val.Kind() == reflect.Interface && !val.IsNil() {
		val = val.Elem()
	}
	return val
}

func CanInline(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Map:
		return !CanExpand(t.Elem())
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if CanExpand(t.Field(i).Type) {
				return false
			}
		}
		return true
	case reflect.Interface:
		return false
	case reflect.Array, reflect.Slice:
		return !CanExpand(t.Elem())
	case reflect.Ptr:
		return false
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return false
	}
	return true
}

func CanExpand(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Map, reflect.Struct,
		reflect.Interface, reflect.Array, reflect.Slice,
		reflect.Ptr:
		return true
	}
	return false
}

func LabelType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Interface, reflect.Struct:
		return true
	}
	return false
}

func RunAllMethod(st interface{}, args ...interface{}) (err error) {
	object := reflect.ValueOf(st)
	kind := object.Kind()
	if kind != reflect.Ptr && kind != reflect.Struct {
		err = errors.New("must be struct")
		return
	}
	for i := 0; i < object.NumMethod(); i++ {
		v := object.Method(i)
		var values []reflect.Value
		for _, v := range args {
			values = append(values, reflect.ValueOf(v))
		}
		v.Call(values)
	}
	return
}
