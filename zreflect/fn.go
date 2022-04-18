package zreflect

import (
	"errors"
	"reflect"
)

func (t *Typer) ForEachMethod(fn func(index int, method reflect.Method, value reflect.Value) error) error {
	v := t.val.Addr()
	numMethod := v.NumMethod()
	if numMethod == 0 {
		return errors.New("method cannot be obtained")
	}
	tp := v.Type()
	var err error
	for i := 0; i < numMethod; i++ {
		err = fn(i, tp.Method(i), v.Method(i))
		if err != nil {
			return err
		}
	}
	return nil
}
