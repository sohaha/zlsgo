package zvalid

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
)

type ele struct {
	target interface{}
	source *Engine
}

// Silent an error occurred during filtering, no error is returned
func (v *Engine) Silent() *Engine {
	v.silent = true
	return v
}

// Default if a filtering error occurs, the default value is assigned to the variable
func (v *Engine) Default(value interface{}) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		v.defaultValue = value
		return v
	})
}

// Separator specify the separator of the slice type
func (v *Engine) Separator(sep string) *Engine {
	if v.err != nil || v.value == "" {
		return v
	}
	v.sep = sep
	return v
}

// Batch assign multiple filtered results to the specified object
func Batch(elements ...*ele) error {
	for k := range elements {
		err := Var(elements[k].target, elements[k].source)
		if err != nil {
			return err
		}
	}
	return nil
}

// BatchVar assign the filtered result to the specified variable
func BatchVar(target interface{}, source *Engine) *ele {
	return &ele{
		target: target,
		source: source,
	}
}

// Var assign the filtered result to the specified variable
func Var(target interface{}, source *Engine) error {
	if source.err != nil && source.defaultValue == nil {
		if source.silent {
			return nil
		}
		return source.err
	}

	targetValueOf := reflect.ValueOf(target)
	if targetValueOf.Kind() != reflect.Ptr {
		if source.silent {
			return nil
		}
		return errors.New(source.name + "参数必须传入指针类型")
	}
	if !targetValueOf.Elem().CanSet() {
		if source.silent {
			return nil
		}
		return errors.New(source.name + "无法更改目标变量的值")
	}
	targetTypeOf := targetValueOf.Elem().Type().Kind()

	if source.err == nil && source.value != "" {
		source.err = setRawValue(targetTypeOf, targetValueOf, source.value, source.sep)
	}

	if source.err != nil && source.defaultValue != nil {
		if err := setDefaultValue(targetTypeOf, targetValueOf, source.defaultValue); err != nil {
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

func setRawValue(targetTypeOf reflect.Kind, targetValueOf reflect.Value, value string, sep string) error {
	typeErr := errors.New("不能用" + targetTypeOf.String() + "类型赋值")
	switch targetTypeOf {
	case reflect.String:
		targetValueOf.Elem().SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return errors.New("必须是整数")
		}
		if targetValueOf.Elem().OverflowInt(v) {
			return typeErr
		}
		targetValueOf.Elem().SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return errors.New("必须是无符号整数")
		}
		if targetValueOf.Elem().OverflowUint(v) {
			return typeErr
		}
		targetValueOf.Elem().SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(value, 10)
		if err != nil {
			return errors.New("必须是小数")
		}
		if targetValueOf.Elem().OverflowFloat(v) {
			return typeErr
		}
	case reflect.Bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return errors.New("必须是布尔值")
		}
		targetValueOf.Elem().SetBool(v)
	case reflect.Slice:
		sliceType := targetValueOf.Elem().Type().String()
		if sliceType == "[]string" {
			if sep == "" {
				return errors.New("过滤规则的分隔符参数(sep)未定义")
			}
			targetValueOf.Elem().Set(reflect.ValueOf(strings.Split(value, sep)))
		}
	default:
		return typeErr
	}

	return nil
}

func setDefaultValue(targetTypeOf reflect.Kind, targetValueOf reflect.Value, value interface{}) error {
	valueTypeOf := reflect.TypeOf(value)
	if valueTypeOf.Kind() != targetTypeOf {
		return errors.New("值类型默认值类型不相同")
	}
	targetValueOf.Elem().Set(reflect.ValueOf(value))
	return nil
}
