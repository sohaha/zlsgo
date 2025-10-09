package ztype

import (
	"errors"
	"reflect"
	"strings"

	"github.com/sohaha/zlsgo/zreflect"
)

type (
	// StruBuilder provides functionality to dynamically build struct types at runtime.
	// It supports creating regular structs, map[T]struct, and []struct types.
	StruBuilder struct {
		key       reflect.Type
		fields    map[string]*StruField
		fieldKeys []string
		typ       int
	}
	// StruField represents a field in a dynamically built struct.
	// It contains the field type and tag information.
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

// NewStructFromValue creates a new StruBuilder from an existing struct value.
// It analyzes the provided struct and copies all its fields to the builder.
func NewStructFromValue(v interface{}) (*StruBuilder, error) {
	b := NewStruct()
	err := b.Merge(v)
	return b, err
}

// NewStruct creates a new StruBuilder for building regular struct types.
func NewStruct() *StruBuilder {
	return &StruBuilder{
		typ:    typeStruct,
		fields: map[string]*StruField{},
	}
}

// NewMapStruct creates a new StruBuilder for building map[T]struct types.
// The key parameter specifies the map key type.
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

// NewSliceStruct creates a new StruBuilder for building []struct types.
func NewSliceStruct() *StruBuilder {
	return &StruBuilder{
		typ:    typeSliceStruct,
		fields: map[string]*StruField{},
	}
}

// Copy copies the configuration from another StruBuilder while preserving the current type.
func (b *StruBuilder) Copy(v *StruBuilder) *StruBuilder {
	typ := b.typ
	val := *v
	*b = val
	b.typ = typ
	return b
}

func (b *StruBuilder) String() string {
	return ToString(b.Interface())
}

// Merge merges fields from one or more struct values into this builder.
// All provided values must be struct types.
func (b *StruBuilder) Merge(values ...interface{}) error {
	for _, value := range values {
		valueOf := reflect.Indirect(zreflect.ValueOf(value))
		typeOf := valueOf.Type()
		if typeOf.Kind() != reflect.Struct {
			return errors.New("value must be struct")
		}

		for i := 0; i < valueOf.NumField(); i++ {
			fval := valueOf.Field(i)
			ftyp := typeOf.Field(i)
			b.AddField(ftyp.Name, fval.Interface(), string(ftyp.Tag))
		}
	}

	return nil
}

// func (b *StruBuilder) AddFunc(name string, fieldType interface{}, tag ...string) *StruBuilder {
// 	reflect.MakeFunc()
// 	return b
// }

// AddField adds a new field to the struct being built.
// The fieldType can be a reflect.Type, another StruBuilder, or any value whose type will be used.
// Optional tag strings will be joined with spaces to form the struct tag.
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

// RemoveField removes a field from the struct being built.
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

// HasField checks if a field with the given name exists in the struct.
func (b *StruBuilder) HasField(name string) bool {
	_, ok := b.fields[name]
	return ok
}

// GetField retrieves a field by name. Returns nil if the field doesn't exist.
func (b *StruBuilder) GetField(name string) *StruField {
	if !b.HasField(name) {
		return nil
	}
	return b.fields[name]
}

// FieldNames returns a slice containing all field names in the struct.
func (b *StruBuilder) FieldNames() []string {
	return b.fieldKeys
}

// Interface returns the built struct as an interface{}.
func (b *StruBuilder) Interface() interface{} {
	return b.Value().Interface()
}

// Type returns the reflect.Type of the built struct.
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

// Value returns a new reflect.Value of the built struct type.
func (b *StruBuilder) Value() reflect.Value {
	return reflect.New(b.Type())
}

// SetType sets the type of this field.
func (f *StruField) SetType(typ interface{}) *StruField {
	f.typ = typ
	return f
}

// SetTag sets the struct tag for this field.
func (f *StruField) SetTag(tag string) *StruField {
	f.tag = tag
	return f
}
