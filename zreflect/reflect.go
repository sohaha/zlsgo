package zreflect

import (
	"errors"
	"reflect"
	"sync"

	"github.com/sohaha/zlsgo/zstring"
)

type Typer struct {
	typ           reflect.Type
	val           reflect.Value
	fieldsMapping map[string]int
	fields        map[string]int
	name          string
	ptr           uintptr
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
	return newVal(val.Elem())
}

func newVal(val reflect.Value) (*Typer, error) {
	ot, err := registerValue(val.Type())
	if err != nil {
		return nil, err
	}
	t := *ot
	t.val = val
	return &t, nil
}

func (t *Typer) Fields() map[string]int {
	if t.fields == nil {

	}
	return t.fields
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

func (t *Typer) Interface() interface{} {
	return reflect.New(t.typ).Interface()
}

func (t *Typer) CheckExistsField(name string) (int, bool) {
	if t.fieldsMapping == nil {
		return CheckExistsField(t.name, name)
	}
	i, ok := t.fieldsMapping[GetFieldTypName(t.name, name)]
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
	t.fieldsMapping = nil
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
	var t *Typer
	can, typName := typ.Name() != "", ""
	if can {
		typName = typ.String()
		if v, ok := registerMap.Load(typName); ok {
			if t, ok = v.(*Typer); ok {
				return t, nil
			}
		}
	}

	t = &Typer{typ: typ, name: typName, fieldsMapping: map[string]int{}, fields: map[string]int{}}

	var register func(typ reflect.Type, name string)
	register = func(r reflect.Type, name string) {
		for i := 0; i < r.NumField(); i++ {
			field := r.Field(i)
			if zstring.IsLcfirst(field.Name) {
				continue
			}
			tag := GetStructTag(field)
			mapFieldName := GetFieldTypName(name, tag)
			typ := field.Type
			if can {
				fieldTagMap.Store(mapFieldName, i)
			}
			t.fieldsMapping[mapFieldName] = i
			t.fields[tag] = i
			if typ.Kind() == reflect.Struct {
				register(typ, mapFieldName)
			}
		}
	}
	register(typ, t.name)
	if can {
		registerMap.Store(t.name, t)
	}
	return t, nil
}
