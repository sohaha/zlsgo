package zdi

import (
	"reflect"
)

type (
	Injector interface {
		Invoker
		TypeMapper
		Set(reflect.Type, reflect.Value)
		Get(reflect.Type) (reflect.Value, bool)
		SetParent(Injector)
	}
	Invoker interface {
		Apply(Pointer) error
		Resolve(...Pointer) error
		Invoke(interface{}) ([]reflect.Value, error)
	}
	TypeMapper interface {
		Map(interface{}, ...Option) reflect.Type
		Maps(...interface{}) []reflect.Type
		Provide(interface{}, ...Option) []reflect.Type
	}
)

type (
	Pointer   interface{}
	Option    func(*mapOption)
	mapOption struct {
		key reflect.Type
	}
	injector struct {
		values    map[reflect.Type]reflect.Value
		providers map[reflect.Type]reflect.Value
		parent    Injector
	}
)

func New(parent ...Injector) Injector {
	inj := &injector{
		values:    make(map[reflect.Type]reflect.Value),
		providers: make(map[reflect.Type]reflect.Value),
	}
	if len(parent) > 0 {
		inj.parent = parent[0]
	}
	return inj
}

func (inj *injector) SetParent(parent Injector) {
	inj.parent = parent
}

func WithInterface(ifacePtr Pointer) Option {
	return func(opt *mapOption) {
		opt.key = ifeOf(ifacePtr)
	}
}

func ifeOf(value interface{}) reflect.Type {
	t := reflect.TypeOf(value)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Interface {
		panic("called inject.key with a value that is not a pointer to an interface. (*MyInterface)(nil)")
	}
	return t
}
