package zreflect

import (
	"fmt"
	"reflect"
)

// GetAllMethod get all methods of struct
func GetAllMethod(s interface{}, fn func(numMethod int, m reflect.Method) error) error {
	typ := ValueOf(s)
	if fn == nil {
		return nil
	}
	return ForEachMethod(typ, func(index int, method reflect.Method, value reflect.Value) error {
		return fn(index, method)
	})
}

// RunAssignMethod run assign methods of struct
func RunAssignMethod(st interface{}, filter func(methodName string) bool, args ...interface{}) (err error) {
	valueOf := ValueOf(st)
	err = GetAllMethod(st, func(numMethod int, m reflect.Method) (err error) {
		if !filter(m.Name) {
			return nil
		}

		var values []reflect.Value
		for _, v := range args {
			values = append(values, ValueOf(v))
		}

		(func() {
			defer func() {
				if e := recover(); e != nil {
					err = fmt.Errorf("method invocation panic: %v", e)
				}
			}()
			valueOf.Method(numMethod).Call(values)
		})()

		return err
	})

	return
}
