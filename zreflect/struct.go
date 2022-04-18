package zreflect

import (
	"errors"
	"reflect"

	"github.com/sohaha/zlsgo/ztime"
	"github.com/sohaha/zlsgo/ztype"
)

var (
	// ErrSkipStruct Field is returned when a struct field is skipped.
	ErrSkipStruct = errors.New("skip struct")
)

func MapToStruct(from map[string]interface{}, obj interface{}) error {
	val := reflect.ValueOf(obj)
	t, err := NewVal(val)
	if err != nil {
		return err
	}

	return MapTypStruct(from, t)
}

func MapTypStruct(from map[string]interface{}, t *Typer) error {
	if t.val == zeroValue {
		return errors.New("no reflect.Value")
	}
	val := t.val
	for n := range from {
		v := from[n]
		index, exists := t.CheckExistsField(n)
		if !exists {
			continue
		}
		field := val.Field(index)
		err := SetStructFidld(t.name, n, field, v)
		if err != nil {
			return err
		}
	}
	return nil
}

// CheckExistsField check field is exists by name
func CheckExistsField(typeName, fieldName string) (index int, exists bool) {
	i, ok := fieldTagMap.Load(typeName + nameConnector + fieldName)
	if !ok {
		return -1, false
	}
	return i.(int), true
}

// ForEachVal For Each Struct field
func (t *Typer) ForEachVal(fn func(parent []string, index int, tag string, field reflect.StructField, val reflect.Value) error) (err error) {
	if t.val == zeroValue {
		return errors.New("no reflect.Value")
	}
	var forField func(t *Typer, v reflect.Value, parent []string) error
	forField = func(t *Typer, v reflect.Value, parent []string) error {
		for i := 0; i < t.typ.NumField(); i++ {
			field := t.Field(i)
			fieldTag := GetStructTag(field)
			if _, ok := t.CheckExistsField(fieldTag); !ok {
				continue
			}
			err = fn(parent, i, fieldTag, field, v.Field(i))
			if err == ErrSkipStruct {
				continue
			}

			if err == nil && field.Type.Kind() == reflect.Struct {
				nt := getTyper()
				nt.typ = field.Type
				nt.fields = t.fields
				nt.name = t.GetFieldTypName(fieldTag)
				err = forField(nt, v.Field(i), append(parent, fieldTag))
				putTyper(nt)
			}

			if err != nil {
				return err
			}
		}
		return nil
	}

	return forField(t, t.val, []string{})
}

// ForEach For Each Struct field
func (t *Typer) ForEach(fn func(parent []string, index int, tag string, field reflect.StructField) error) (err error) {
	var forField func(t *Typer, parent []string) error
	forField = func(t *Typer, parent []string) error {
		for i := 0; i < t.typ.NumField(); i++ {
			field := t.Field(i)
			fieldTag := GetStructTag(field)
			if _, ok := t.CheckExistsField(fieldTag); !ok {
				continue
			}

			err = fn(parent, i, fieldTag, field)
			if err == ErrSkipStruct {
				continue
			}

			if err == nil && field.Type.Kind() == reflect.Struct {
				nt := getTyper()
				nt.typ = field.Type
				nt.fields = t.fields
				nt.name = t.GetFieldTypName(fieldTag)
				err = forField(nt, append(parent, fieldTag))
				putTyper(nt)
			}

			if err != nil {
				return err
			}
		}
		return nil
	}

	return forField(t, []string{})
}

func SetStructFidld(typName, tag string, fValue reflect.Value, val interface{}) error {
	tp := fValue.Type()
	fkind := tp.Kind()
	if !fValue.CanSet() {
		return nil
	}
	switch fkind {
	case reflect.Bool:
		if val == nil {
			fValue.SetBool(false)
		} else if v, ok := val.(bool); ok {
			fValue.SetBool(v)
		} else {
			v := ztype.ToBool(val)
			fValue.SetBool(v)
		}
	case reflect.String:
		s, ok := val.(string)
		if !ok {
			s = ztype.ToString(val)
		}
		fValue.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fValue.SetInt(ztype.ToInt64(val))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fValue.SetUint(ztype.ToUint64(val))
	case reflect.Float64, reflect.Float32:
		fValue.SetFloat(ztype.ToFloat64(val))
	case reflect.Struct:
		if val == nil {
			fValue.Set(reflect.Zero(tp))
			return nil
		} else if vmap, ok := val.(map[string]interface{}); ok {
			t := &Typer{name: typName + nameConnector + tag, typ: tp, val: fValue}
			return MapTypStruct(vmap, t)
		} else if isTimeType(fValue) {
			var (
				timeString string
				timeInt    int64
			)
			switch d := val.(type) {
			case []byte:
				timeString = string(d)
			case string:
				timeString = d
			case int64:
				timeInt = d
			case int:
				timeInt = int64(d)
			}
			if timeInt > 0 {
				fValue.Set(reflect.ValueOf(ztime.Unix(timeInt)))
			} else if timeString != "" {
				t, err := ztime.Parse(timeString)
				if err == nil {
					fValue.Set(reflect.ValueOf(t))
				}
			}
		} else {
			valVof := reflect.ValueOf(val)
			if valVof.Type() == tp {
				fValue.Set(valVof)
			} else if valVof.Kind() == reflect.Map {
				nv := make(map[string]interface{}, valVof.Len())
				mapKeys := valVof.MapKeys()
				for i := range mapKeys {
					m := mapKeys[i]
					nv[m.String()] = valVof.MapIndex(m).Interface()
				}
				t := &Typer{name: typName + nameConnector + tag, typ: tp, val: fValue}
				return MapTypStruct(nv, t)
			}
			// return errors.New("not support " + fkind.String())
		}
	default:
		v := reflect.ValueOf(val)
		if v.Type() == tp {
			fValue.Set(v)
		}
	}
	return nil
}
