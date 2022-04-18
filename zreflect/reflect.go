package zreflect

import (
	"errors"
	"reflect"
	"sync"

	"github.com/sohaha/zlsgo/zstring"
)

type Typer struct {
	typ    reflect.Type
	val    reflect.Value
	fields map[string]int
	name   string
	ptr    uintptr
}

var (
	Tag = "z"
)

var (
	zeroValue   = reflect.Value{}
	fieldTagMap sync.Map
	registerMap sync.Map
)

func TypeOf(obj interface{}) reflect.Type {
	return getTypElem(reflect.TypeOf(obj))
}

func ValueOf(obj interface{}) (v reflect.Value, err error) {
	v = reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr {
		err = errors.New("not ptr")
	}
	return
}

func NewTyp(typ reflect.Type) (*Typer, error) {
	return registerValue(typ)
}

func NewVal(val reflect.Value) (*Typer, error) {
	if val.Kind() != reflect.Ptr {
		return nil, errors.New("not ptr")
	}
	ptr := val.Pointer()
	val = val.Elem()
	t, err := registerValue(val.Type())
	if err != nil {
		return nil, err
	}
	t.val = val
	t.ptr = ptr
	return t, nil
}

func (t *Typer) Field(i int) reflect.StructField {
	return t.typ.Field(i)
}

func (t *Typer) ValueOf() reflect.Value {
	return t.val
}

func (t *Typer) TypeOf() reflect.Type {
	return t.typ
}

func (t *Typer) CheckExistsField(name string) (int, bool) {
	if t.fields == nil {
		return CheckExistsField(t.name, name)
	}
	i, ok := t.fields[GetFieldTypName(t.name, name)]
	return i, ok
}

func (t *Typer) GetFieldTypName(name string) string {
	return GetFieldTypName(t.name, name)
}

var pool = sync.Pool{New: func() interface{} {
	return &Typer{}
}}

func getTyper() *Typer {
	return pool.Get().(*Typer)
}

func putTyper(t *Typer) {
	t.typ = nil
	t.fields = nil
	t.name = ""
	pool.Put(t)
}

func (t *Typer) Name() string {
	return t.name
}

func Register(obj interface{}) error {
	typ, ok := obj.(reflect.Type)
	if !ok {
		typ = reflect.TypeOf(obj)
	}
	_, err := registerValue(typ)
	return err
}

func GetFieldTypName(typName, fieldName string) string {
	return typName + nameConnector + fieldName
}

func getTypElem(typ reflect.Type) reflect.Type {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ
}

func registerValue(typ reflect.Type) (*Typer, error) {
	typ = getTypElem(typ)
	if typ.Kind() != reflect.Struct {
		return nil, errors.New("only registered structure")
	}
	can := typ.Name() != ""
	var t *Typer
	if can {
		typName := typ.String()
		if _, ok := registerMap.Load(typName); ok {
			return &Typer{typ: typ, name: typName}, nil
		}
		t = &Typer{typ: typ, name: typName}
	} else {

		t = &Typer{typ: typ, name: "", fields: map[string]int{}}
	}

	var register func(typ reflect.Type, name string)
	register = func(r reflect.Type, name string) {
		for i := 0; i < r.NumField(); i++ {
			field := r.Field(i)
			if zstring.IsLcfirst(field.Name) {
				continue
			}
			mapFieldName := GetFieldTypName(name, GetStructTag(field))
			typ := field.Type
			if can {
				mapFieldName = GetFieldTypName(name, GetStructTag(field))
				fieldTagMap.Store(mapFieldName, i)
			} else {
				t.fields[mapFieldName] = i
			}
			if typ.Kind() == reflect.Struct {
				register(typ, mapFieldName)
			}
		}
	}
	register(typ, t.name)
	if can {
		registerMap.Store(t.name, struct{}{})
	}
	return t, nil
}
