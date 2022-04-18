// Package zvalid data verification
package zvalid

import (
	"container/list"
	"errors"
	"strconv"
)

type (
	// Engine valid engine
	Engine struct {
		queue        *list.List
		setRawValue  bool
		err          error
		name         string
		value        string
		valueInt     int
		valueFloat   float64
		sep          string
		silent       bool
		defaultValue interface{}
		result       bool
	}
	queueT func(v *Engine) *Engine
)

var (
	// ErrNoValidationValueSet no verification value set
	ErrNoValidationValueSet = errors.New("未设置验证值")
)

// New  valid
func New() Engine {
	return Engine{
		queue: list.New(),
	}
}

// Int use int new valid
func Int(value int, name ...string) Engine {
	return Text(strconv.FormatInt(int64(value), 10), name...)
}

// Text use int new valid
func Text(value string, name ...string) Engine {
	var obj Engine
	obj.value = value
	obj.setRawValue = true
	obj.queue = list.New()
	if len(name) > 0 {
		obj.name = name[0]
	}
	return obj
}

// Required Must have a value (zero values ​​other than "" are allowed). If this rule is not used, when the parameter value is "", data validation does not take effect by default
func (v Engine) Required(customError ...string) Engine {
	return pushQueue(&v, func(v *Engine) *Engine {
		if v.value == "" {
			v.err = setError(v, "不能为空", customError...)
		}
		return v
	})
}

// Customize customize valid
func (v Engine) Customize(fn func(rawValue string, err error) (newValue string, newErr error)) Engine {
	return pushQueue(&v, func(v *Engine) *Engine {
		v.value, v.err = fn(v.value, v.err)
		return v
	}, true)
}

func pushQueue(v *Engine, fn queueT, DisableCheckErr ...bool) Engine {
	pFn := fn
	if !(len(DisableCheckErr) > 0 && DisableCheckErr[0]) {
		pFn = func(v *Engine) *Engine {
			if v.err != nil {
				return v
			}
			return fn(v)
		}
	}
	queue := list.New()
	queue.PushBackList(v.queue)
	queue.PushBack(pFn)
	v.queue = queue
	return *v
}

func ignore(v *Engine) bool {
	return v.err != nil || v.value == ""
}

func notEmpty(v *Engine) bool {
	return v.value != ""
}

func setError(v *Engine, msg string, customError ...string) error {
	if len(customError) > 0 {
		return errors.New(customError[0])
	}
	if v.name != "" {
		msg = v.name + msg
	}

	return errors.New(msg)
}
