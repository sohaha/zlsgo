package ztype

import (
	"reflect"
	"strings"
)

type (
	StructBuilder struct {
		fields map[string]*StructField
		typ    int
		key    interface{}
	}
	StructField struct {
		typ interface{}
		tag string
	}
)

const (
	typeStruct = iota
	typeMapStruct
	typeSliceStruct
)

func NewStruct() *StructBuilder {
	return &StructBuilder{
		typ:    typeStruct,
		fields: map[string]*StructField{},
	}
}

func NewMapStruct(key interface{}) *StructBuilder {
	return &StructBuilder{
		typ:    typeMapStruct,
		key:    key,
		fields: map[string]*StructField{},
	}
}

func NewtSliceStruct() *StructBuilder {
	return &StructBuilder{
		typ:    typeSliceStruct,
		fields: map[string]*StructField{},
	}
}

func (b *StructBuilder) AddField(name string, fieldType interface{}, tag ...string) *StructBuilder {
	var t string
	if len(tag) > 0 {
		t = strings.Join(tag, " ")
	}
	b.fields[name] = &StructField{
		typ: fieldType,
		tag: t,
	}
	return b
}

func (b *StructBuilder) RemoveField(name string) *StructBuilder {
	delete(b.fields, name)

	return b
}

func (b *StructBuilder) HasField(name string) bool {
	_, ok := b.fields[name]
	return ok
}

func (b *StructBuilder) GetField(name string) *StructField {
	if !b.HasField(name) {
		return nil
	}
	return b.fields[name]
}

func (b *StructBuilder) Interface() interface{} {
	return b.Value().Interface()
}

func (b *StructBuilder) Value() reflect.Value {
	var structFields []reflect.StructField
	for name, field := range b.fields {
		t, ok := field.typ.(reflect.Type)
		if !ok {
			t = reflect.TypeOf(field.typ)
		}
		structFields = append(structFields, reflect.StructField{
			Name: name,
			Type: t,
			Tag:  reflect.StructTag(field.tag),
		})
	}
	v := reflect.StructOf(structFields)
	switch b.typ {
	case typeStruct:
		return reflect.New(v)
	case typeMapStruct:
		return reflect.New(reflect.MapOf(reflect.Indirect(reflect.ValueOf(b.key)).Type(), v))
	default:
		return reflect.New(reflect.SliceOf(v))
	}
}

func (f *StructField) SetType(typ interface{}) *StructField {
	f.typ = typ
	return f
}

func (f *StructField) SetTag(tag string) *StructField {
	f.tag = tag
	return f
}
