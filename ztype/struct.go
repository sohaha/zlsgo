package ztype

import (
	"reflect"
	"strings"

	"github.com/sohaha/zlsgo/zreflect"
)

type (
	StruBuilder struct {
		key       reflect.Type
		fields    map[string]*StruField
		fieldKeys []string
		typ       int
	}
	StruField struct {
		typ interface{}
		tag string
	}
)

const (
	typeStruct = iota
	typeMapStruct
	typeSliceStruct
)

func NewStruct() *StruBuilder {
	return &StruBuilder{
		typ:    typeStruct,
		fields: map[string]*StruField{},
	}
}

func NewMapStruct(key interface{}) *StruBuilder {
	var k reflect.Type
	if v, ok := key.(reflect.Type); ok {
		k = v
	} else {
		k = reflect.TypeOf(key)
	}
	return &StruBuilder{
		typ:    typeMapStruct,
		key:    k,
		fields: map[string]*StruField{},
	}
}

func NewSliceStruct() *StruBuilder {
	return &StruBuilder{
		typ:    typeSliceStruct,
		fields: map[string]*StruField{},
	}
}

func (b *StruBuilder) Copy(v *StruBuilder) *StruBuilder {
	typ := b.typ
	val := *v
	*b = val
	b.typ = typ
	return b
}

func (b *StruBuilder) Merge(values ...interface{}) *StruBuilder {
	for _, value := range values {
		valueOf := reflect.Indirect(zreflect.ValueOf(value))
		typeOf := valueOf.Type()
		for i := 0; i < valueOf.NumField(); i++ {
			fval := valueOf.Field(i)
			ftyp := typeOf.Field(i)
			b.AddField(ftyp.Name, fval.Interface(), string(ftyp.Tag))
		}
	}

	return b
}

// func (b *StruBuilder) AddFunc(name string, fieldType interface{}, tag ...string) *StruBuilder {
// 	reflect.MakeFunc()
// 	return b
// }

func (b *StruBuilder) AddField(name string, fieldType interface{}, tag ...string) *StruBuilder {
	var t string
	if len(tag) > 0 {
		t = strings.Join(tag, " ")
	}
	if b.typ == typeStruct {
		nkey := make([]string, 0, len(b.fieldKeys))
		for i := range b.fieldKeys {
			if b.fieldKeys[i] != name {
				nkey = append(nkey, b.fieldKeys[i])
			}
		}
		b.fieldKeys = append(nkey, name)
	}
	b.fields[name] = &StruField{
		typ: fieldType,
		tag: t,
	}
	return b
}

func (b *StruBuilder) RemoveField(name string) *StruBuilder {
	delete(b.fields, name)
	if b.typ == typeStruct {
		nkey := make([]string, 0, len(b.fieldKeys))
		for i := range b.fieldKeys {
			if b.fieldKeys[i] != name {
				nkey = append(nkey, b.fieldKeys[i])
			}
		}
		b.fieldKeys = nkey
	}
	return b
}

func (b *StruBuilder) HasField(name string) bool {
	_, ok := b.fields[name]
	return ok
}

func (b *StruBuilder) GetField(name string) *StruField {
	if !b.HasField(name) {
		return nil
	}
	return b.fields[name]
}

func (b *StruBuilder) Interface() interface{} {
	return b.Value().Interface()
}

func (b *StruBuilder) Type() reflect.Type {
	var fields []reflect.StructField
	fn := func(name string, field *StruField) {
		var t reflect.Type
		switch v := field.typ.(type) {
		case *StruBuilder:
			t = v.Type()
		case reflect.Type:
			t = v
		default:
			t = reflect.TypeOf(field.typ)
		}

		fields = append(fields, reflect.StructField{
			Name: name,
			Type: t,
			Tag:  reflect.StructTag(field.tag),
		})
	}
	if b.typ == typeStruct {
		for i := range b.fieldKeys {
			name := b.fieldKeys[i]
			if field, ok := b.fields[name]; ok {
				fn(name, field)
			}
		}
	} else {
		for name := range b.fields {
			field := b.fields[name]
			fn(name, field)
		}
	}

	typ := reflect.StructOf(fields)

	switch b.typ {
	case typeSliceStruct:
		return reflect.SliceOf(typ)
	case typeMapStruct:
		return reflect.MapOf(b.key, typ)
	default:
		return typ
	}
}

func (b *StruBuilder) Value() reflect.Value {
	return reflect.New(b.Type())
}

func (f *StruField) SetType(typ interface{}) *StruField {
	f.typ = typ
	return f
}

func (f *StruField) SetTag(tag string) *StruField {
	f.tag = tag
	return f
}
