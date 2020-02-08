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
