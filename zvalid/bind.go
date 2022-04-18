package zvalid

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// ValidEle ValidEle
type ValidEle struct {
	target interface{}
	source Engine
}

// Silent an error occurred during filtering, no error is returned
func (v Engine) Silent() Engine {
	v.silent = true
	return v
}

// Default if a filtering error occurs, the default value is assigned to the variable
func (v Engine) Default(value interface{}) Engine {
	return pushQueue(&v, func(v *Engine) *Engine {
		v.defaultValue = value
		return v
	}, true)
}

// Separator specify the separator of the slice type
func (v Engine) Separator(sep string) Engine {
	if v.err != nil || v.value == "" {
		return v
	}
	v.sep = sep
	return v
}

// Batch assign multiple filtered results to the specified object
func Batch(elements ...*ValidEle) error {
	for k := range elements {
		e := elements[k]
		if e == nil {
			return nil
		}
		err := Var(e.target, e.source)
		if err != nil {
			return err
		}
	}
	return nil
}

// BatchVar assign the filtered result to the specified variable
func BatchVar(target interface{}, source Engine) *ValidEle {
	return &ValidEle{
		target: target,
		source: source,
	}
}

// Var assign the filtered result to the specified variable
func Var(target interface{}, source Engine, name ...string) error {
	source = *source.valid()
	if len(name) > 0 {
		source.name = name[0]
	}
	if source.err != nil && source.defaultValue == nil {
		if source.silent {
			return nil
		}
		return source.err
	}
	var (
		val reflect.Value
		k   reflect.Kind
	)
	val, ok := target.(reflect.Value)
	if !ok {
		val = reflect.ValueOf(target)
		if val.Kind() != reflect.Ptr {
			if source.silent {
				return nil
			}
			return fmt.Errorf("parameter must pass in a pointer type: %s", source.name)
		}
		if !val.Elem().CanSet() {
			if source.silent {
				return nil
			}
			return fmt.Errorf("target value of the variable cannot be changed: %s", source.name)
		}
		val = val.Elem()
		k = val.Type().Kind()
	} else {
		k = val.Type().Kind()
	}

	if source.err == nil && source.value != "" {
		source.err = setRawValue(k, val, source.value, source.sep)
	}

	if source.err != nil && source.defaultValue != nil {
		if err := setDefaultValue(k, val, source.defaultValue); err != nil {
			if source.silent {
				return nil
			}
			return err
		}
		return nil
	} else if source.err != nil {
		if source.silent {
			return nil
		}
		return errors.New(source.name + source.err.Error())
	}

	return nil
}

func setRawValue(k reflect.Kind, val reflect.Value, value string, sep string) error {
	typeErr := errors.New("不能用" + k.String() + "类型赋值")
	switch k {
	case reflect.String:
		val.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return errors.New("必须是整数")
		}
		if val.OverflowInt(v) {
			return typeErr
		}
		val.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return errors.New("必须是无符号整数")
		}
		if val.OverflowUint(v) {
			return typeErr
		}
		val.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return errors.New("必须是小数")
		}
		if val.OverflowFloat(v) {
			return typeErr
		}
		val.SetFloat(v)
	case reflect.Bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return errors.New("必须是布尔值")
		}
		val.SetBool(v)
	case reflect.Slice:
		sliceType := val.Type().String()
		if sliceType == "[]string" {
			if sep == "" {
				return errors.New("过滤规则的分隔符参数(sep)未定义")
			}
			val.Set(reflect.ValueOf(strings.Split(value, sep)))
		}
	default:
		return typeErr
	}

	return nil
}

func setDefaultValue(targetTypeOf reflect.Kind, targetValueOf reflect.Value, value interface{}) error {
	valueTypeOf := reflect.ValueOf(value)
	if valueTypeOf.Kind() != targetTypeOf {
		return errors.New("值类型默认值类型不相同" + valueTypeOf.String() + "/" + targetTypeOf.String())
	}
	targetValueOf.Set(valueTypeOf)
	return nil
}
