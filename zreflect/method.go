package zreflect

import (
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
	err = GetAllMethod(st, func(numMethod int, m reflect.Method) error {
		if !filter(m.Name) {
			return nil
		}

		var values []reflect.Value
		for _, v := range args {
			values = append(values, ValueOf(v))
		}

		err := func() error {
			defer func() {
				if e := recover(); e != nil {
					err, _ = e.(error)
				}
			}()
			valueOf.Method(numMethod).Call(values)
			return nil
		}

		return err()
	})

	return
}
