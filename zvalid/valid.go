package zvalid

import (
	"container/list"
	"errors"
)

type (
	Engine struct {
		queue        *list.List
		setRawValue  bool
		err          error
		name         string
		value        string
		i            int
		f            float64
		sep          string
		silent       bool
		defaultValue interface{}
	}
	queueT func(v *Engine) *Engine
)

var (
	ErrNoValidationValueSet = errors.New("未设置验证值")
)

func New() *Engine {
	return &Engine{
		queue: list.New(),
	}
}

func clone(v *Engine) *Engine {
	if v.setRawValue {
		return v
	}
	queue := list.New()
	if v.queue.Len() > 0 {
		queue.PushBackList(v.queue)
	}

	return &Engine{
		setRawValue:  v.setRawValue,
		err:          v.err,
		name:         v.name,
		value:        v.value,
		i:            v.i,
		f:            v.f,
		sep:          v.sep,
		silent:       v.silent,
		defaultValue: v.defaultValue,
		queue:        queue,
	}
}

func Text(value string, name ...string) *Engine {
	var obj Engine
	obj.value = value
	obj.setRawValue = true
	if len(name) > 0 {
		obj.name = name[0]
	}
	return &obj
}

func (v *Engine) setError(msg string, customError ...string) error {
	if len(customError) > 0 {
		return errors.New(customError[0])
	}
	return errors.New(v.name + msg)
}

// Required Must have a value (zero values ​​other than "" are allowed). If this rule is not used, when the parameter value is "", data validation does not take effect by default
func (v *Engine) Required(customError ...string) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		if v.err != nil {
			return v
		}
		if v.value == "" {
			v.err = v.setError("不能为空", customError...)
			return v
		}
		return v

	})
}

// Customize customize valid
func (v *Engine) Customize(fn func(rawValue string, err error) (newValue string, newErr error)) *Engine {
	return pushQueue(v, func(v *Engine) *Engine {
		v.value, v.err = fn(v.value, v.err)
		return v
	})
}

func pushQueue(v *Engine, fn queueT) *Engine {
	vc := clone(v)
	vc.queue.PushBack(fn)
	return vc
}

func ignore(v *Engine) bool {
	return v.err != nil || v.value == ""
}

func notEmpty(v *Engine) bool {
	return v.value != ""
}
