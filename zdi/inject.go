package zdi

import (
	"fmt"
	"reflect"

	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zreflect"
)

func (inj *injector) InvokeWithErrorOnly(f interface{}) (err error) {
	v, err := inj.Invoke(f)
	if err != nil {
		return err
	}

	if len(v) == 0 {
		return nil
	}

	for i := range v {
		if err, ok := v[i].Interface().(error); ok {
			return err
		}
	}

	return nil
}

func (inj *injector) Invoke(f interface{}) (values []reflect.Value, err error) {
	catch := zerror.TryCatch(func() error {
		t := zreflect.TypeOf(f)
		switch v := f.(type) {
		case PreInvoker:
			values, err = inj.fast(v, t, t.NumIn())
		default:
			values, err = inj.call(f, t, t.NumIn())
		}
		return nil
	})

	if catch != nil {
		err = catch
	}

	return
}

func (inj *injector) call(f interface{}, t reflect.Type, numIn int) ([]reflect.Value, error) {
	var in []reflect.Value
	if numIn > 0 {
		in = make([]reflect.Value, numIn)
		var argType reflect.Type
		for i := 0; i < numIn; i++ {
			argType = t.In(i)
			val, ok := inj.Get(argType)
			if !ok {
				return nil, fmt.Errorf("value not found for type %v", argType)
			}

			in[i] = val
		}
	}
	return zreflect.ValueOf(f).Call(in), nil
}

func (inj *injector) Map(val interface{}, opt ...Option) (override reflect.Type) {
	o := mapOption{}
	for _, opt := range opt {
		opt(&o)
	}
	if o.key == nil {
		o.key = reflect.TypeOf(val)
	}
	if _, ok := inj.values[o.key]; ok {
		override = o.key
	}

	inj.values[o.key] = zreflect.ValueOf(val)
	return
}

func (inj *injector) Maps(values ...interface{}) (override []reflect.Type) {
	for _, val := range values {
		o := inj.Map(val)
		if o != nil {
			override = append(override, o)
		}
	}
	return
}

func (inj *injector) Set(typ reflect.Type, val reflect.Value) {
	inj.values[typ] = val
}

func (inj *injector) Get(t reflect.Type) (reflect.Value, bool) {
	val := inj.values[t]
	if val.IsValid() {
		return val, true
	}

	if provider, ok := inj.providers[t]; ok {
		results, err := inj.Invoke(provider.Interface())
		if err != nil {
			panic(err)
		}
		for _, result := range results {
			resultType := result.Type()
			inj.values[resultType] = result
			delete(inj.providers, resultType)
			if resultType == t {
				val = result
			}
		}

		if val.IsValid() {
			return val, true
		}
	}

	if t.Kind() == reflect.Interface {
		for k, v := range inj.values {
			if k.Implements(t) {
				val = v
				break
			}
		}
	}

	if val.IsValid() {
		return val, true
	}

	var ok bool
	if inj.parent != nil {
		val, ok = inj.parent.Get(t)
	}

	return val, ok
}
