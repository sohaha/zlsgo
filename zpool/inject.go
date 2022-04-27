package zpool

import (
	"reflect"

	"github.com/sohaha/zlsgo/zdi"
)

type (
	invokerPre func() error
)

func (i invokerPre) Invoke(_ []interface{}) ([]reflect.Value, error) {
	return nil, i()
}

var (
	_ zdi.PreInvoker = (*invokerPre)(nil)
)

func invokeHandler(v []reflect.Value, err error) error {
	if err != nil {
		return err
	}
	for i := range v {
		val := v[i].Interface()
		switch e := val.(type) {
		case error:
			return e
		}
	}
	return nil
}

func (wp *WorkPool) Injector() zdi.TypeMapper {
	return wp.injector
}

func (wp *WorkPool) handlerFunc(h Task) (fn taskfn) {
	switch v := h.(type) {
	case func():
		return func() error {
			v()
			return nil
		}
	case func() error:
		return func() error {
			return invokeHandler(wp.injector.Invoke(invokerPre(v)))
		}
	case zdi.PreInvoker:
		return func() error {
			err := invokeHandler(wp.injector.Invoke(v))
			return err
		}
	default:
		return func() error {
			return invokeHandler(wp.injector.Invoke(v))
		}
	}
}
