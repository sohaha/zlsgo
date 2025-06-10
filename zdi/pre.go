package zdi

import (
	"fmt"
	"reflect"
)

type PreInvoker interface {
	Invoke([]interface{}) ([]reflect.Value, error)
}

// IsPreInvoker checks if the given handler implements the PreInvoker interface.
// PreInvoker allows for a potentially faster invocation path by bypassing some reflection.
func IsPreInvoker(handler interface{}) bool {
	_, ok := handler.(PreInvoker)
	return ok
}

// fast is an internal helper for invoking a PreInvoker.
// It resolves dependencies for the PreInvoker's arguments as interface{} slices
// and then calls its Invoke method.
func (inj *injector) fast(f PreInvoker, t reflect.Type, numIn int) ([]reflect.Value, error) {
	var in []interface{}
	if numIn > 0 {
		in = make([]interface{}, numIn)
		var argType reflect.Type
		for i := 0; i < numIn; i++ {
			argType = t.In(i)
			val, ok := inj.Get(argType)
			if !ok {
				return nil, fmt.Errorf("value not found for type %v", argType)
			}

			in[i] = val.Interface()
		}
	}
	return f.Invoke(in)
}
