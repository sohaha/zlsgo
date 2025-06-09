// Package zdi provides dependency injection
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
		InvokeWithErrorOnly(interface{}) error
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

// New creates and returns a new Injector.
// Optionally, a parent Injector can be provided to enable hierarchical dependency resolution.
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

// SetParent sets the parent Injector for the current injector.
// This allows for chaining injectors, enabling a hierarchical lookup for dependencies.
func (inj *injector) SetParent(parent Injector) {
	inj.parent = parent
}

// WithInterface is an option used with Map or Provide methods.
// It specifies the interface type that a concrete type should be mapped to.
// The argument must be a pointer to an interface type, e.g., (*MyInterface)(nil).
func WithInterface(ifacePtr Pointer) Option {
	return func(opt *mapOption) {
		opt.key = ifeOf(ifacePtr)
	}
}

// ifeOf returns the underlying reflect.Type of an interface pointer.
// It is used internally to extract the interface type for mapping.
// Panics if the provided value is not a pointer to an interface.
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
