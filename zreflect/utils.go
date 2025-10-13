package zreflect

import (
	"reflect"
	"strings"
	_ "unsafe"
)

const (
	Tag            = "z"
	ignoreTagValue = "-"
	nameConnector  = "::"
)

// GetStructTag extracts the tag value from a struct field.
// It first looks for tags provided in the tags parameter. If none are provided or found,
// it falls back to the default "z" tag, then tries the "json" tag.
// If a tag value is "-", it returns empty strings.
// The function also parses tag options (after a comma in the tag value).
func GetStructTag(field reflect.StructField, tags ...string) (tagValue, tagOpts string) {
	if len(tags) > 0 {
		for i := range tags {
			tagValue = field.Tag.Get(tags[i])
			if tagValue == ignoreTagValue {
				return "", ""
			}
			if t, v := checkTagValidity(tagValue); t != "" {
				return t, v
			}
		}
	} else {
		tagValue = field.Tag.Get(Tag)
		if tagValue == ignoreTagValue {
			return "", ""
		}
		if t, v := checkTagValidity(tagValue); t != "" {
			return t, v
		}

		tagValue = field.Tag.Get("json")
		if tagValue == ignoreTagValue {
			return "", ""
		}
		if t, v := checkTagValidity(tagValue); t != "" {
			return t, v
		}
	}

	return field.Name, ""
}

// Nonzero determines whether a reflect.Value is the zero value for its type.
// This is useful for checking if a value has been initialized or set to a non-default value.
func Nonzero(v reflect.Value) bool {
	switch k := v.Kind(); k {
	case reflect.Bool:
		return v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return v.Float() != 0
	case reflect.Complex64, reflect.Complex128:
		return v.Complex() != complex(0, 0)
	case reflect.String:
		return v.String() != ""
	case reflect.Map, reflect.Interface, reflect.Slice, reflect.Ptr, reflect.Chan, reflect.Func:
		return !v.IsNil()
	case reflect.UnsafePointer:
		return v.Pointer() != 0
	case reflect.Struct:
		numField := v.NumField()
		for i := 0; i < numField; i++ {
			field := v.Field(i)
			if field.Kind() == reflect.Ptr {
				if field.IsNil() {
					continue
				}
				field = field.Elem()
			}
			if Nonzero(field) {
				return true
			}
		}
		return false
	case reflect.Array:
		vLen := v.Len()
		for i := 0; i < vLen; i++ {
			if Nonzero(v.Index(i)) {
				return true
			}
		}
		return false
	}
	return true
}

// CanExpand checks if a type can be expanded (e.g., a struct that can be broken down into fields).
// This is used to determine if a value should be displayed as a single item or expanded into its components.
func CanExpand(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Map, reflect.Struct,
		reflect.Interface, reflect.Array, reflect.Slice,
		reflect.Ptr:
		return true
	}
	return false
}

// CanInline determines if a type can be displayed inline rather than requiring a multi-line representation.
// This is useful for formatting and displaying values in a compact way when possible.
func CanInline(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Map:
		return !CanExpand(t.Elem())
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if CanExpand(t.Field(i).Type) {
				return false
			}
		}
		return true
	case reflect.Interface:
		return false
	case reflect.Array, reflect.Slice:
		return !CanExpand(t.Elem())
	case reflect.Ptr:
		return false
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return false
	}
	return true
}

// IsLabel checks if a type should be treated as a label (e.g., a string or simple scalar type).
// This is used for display and formatting purposes.
func IsLabel(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Interface, reflect.Struct:
		return true
	}
	return false
}

// checkTagValidity parses a tag string into its main value and options.
// It handles the common format "value,option1,option2" used in struct tags.
func checkTagValidity(tagValue string) (tag, tagOpts string) {
	if tagValue == "" {
		return "", ""
	}
	valueSplit := strings.SplitN(tagValue, ",", 2)
	if len(valueSplit) < 2 {
		return valueSplit[0], ""
	}
	return valueSplit[0], valueSplit[1]
}

//go:linkname ifaceIndir reflect.ifaceIndir
//go:noescape
func ifaceIndir(Type) bool

// GetAbbrKind returns the abbreviated kind of a reflect.Value.
// It unwraps pointers and interfaces to get to the underlying concrete type.
func GetAbbrKind(val reflect.Value) reflect.Kind {
	kind := val.Kind()
	switch {
	case kind >= reflect.Int && kind <= reflect.Int64:
		return reflect.Int
	case kind >= reflect.Uint && kind <= reflect.Uint64:
		return reflect.Uint
	case kind >= reflect.Float32 && kind <= reflect.Float64:
		return reflect.Float64
	default:
		return kind
	}
}
