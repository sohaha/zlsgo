// Package zdi provides dependency injection
package zdi

import (
	"reflect"

	"github.com/sohaha/zlsgo/zsync"
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
		key  reflect.Type
		name string
	}
	injector struct {
		mu        *zsync.RBMutex
		values    map[reflect.Type]reflect.Value
		providers map[reflect.Type]reflect.Value
		// providerErrIdx records, for each type produced by a provider, the index
		// of the error return value (-1 when the provider has none).
		// error itself is never registered as an injectable key.
		providerErrIdx map[reflect.Type]int
		named          map[string]reflect.Value
		namedProviders map[string]reflect.Value
		inflight       map[string]*inflightCall
		parent         Injector
	}
)

// errorType is the reflect.Type of the error interface, used to detect the
// error return value of a constructor.
var errorType = reflect.TypeOf((*error)(nil)).Elem()

// depStack tracks the keys already entered for instantiation along the current
// resolution chain, used for circular dependency detection. Both type-based and
// name-based dependencies share the same key space, so a provider that
// re-enters itself is reported as an error instead of blocking forever on the
// non-reentrant singleflight key.
type depStack []string

func typeKey(t reflect.Type) string {
	return "type:" + t.PkgPath() + ":" + t.String()
}

func namedKey(name string) string {
	return "name:" + name
}

func (s depStack) contains(key string) bool {
	for i := range s {
		if s[i] == key {
			return true
		}
	}
	return false
}

func (s depStack) push(key string) depStack {
	return append(s, key)
}

// New creates and returns a new Injector.
// Optionally, a parent Injector can be provided to enable hierarchical dependency resolution.
func New(parent ...Injector) Injector {
	inj := &injector{
		values:         make(map[reflect.Type]reflect.Value),
		providers:      make(map[reflect.Type]reflect.Value),
		providerErrIdx: make(map[reflect.Type]int),
		named:          make(map[string]reflect.Value),
		namedProviders: make(map[string]reflect.Value),
		inflight:       make(map[string]*inflightCall),
		mu:             zsync.NewRBMutex(),
	}
	if len(parent) > 0 {
		inj.parent = parent[0]
	}
	return inj
}

// SetParent sets the parent Injector for the current injector.
// This allows for chaining injectors, enabling a hierarchical lookup for dependencies.
func (inj *injector) SetParent(parent Injector) {
	inj.mu.Lock()
	inj.parent = parent
	inj.mu.Unlock()
}

// WithInterface is an option used with Map or Provide methods.
// It specifies the interface type that a concrete type should be mapped to.
// The argument must be a pointer to an interface type, e.g., (*MyInterface)(nil).
func WithInterface(ifacePtr Pointer) Option {
	return func(opt *mapOption) {
		opt.key = ifeOf(ifacePtr)
	}
}

// Name is an option used with Map or Provide methods.
// It registers the value or provider under a name, retrievable via GetNamed
// or the generic ResolveNamed helpers.
func Name(name string) Option {
	return func(opt *mapOption) {
		opt.name = name
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

// resolveOption folds the given options into a single mapOption.
func resolveOption(opts []Option) mapOption {
	o := mapOption{}
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	return o
}
