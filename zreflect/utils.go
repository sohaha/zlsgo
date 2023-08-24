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

func Nonzero(v reflect.Value) bool {
	switch v.Kind() {
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
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if Nonzero(reflect.Indirect(v.Field(i))) {
				return true
			}
		}
		return false
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if Nonzero(v.Index(i)) {
				return true
			}
		}
		return false
	case reflect.Map, reflect.Interface, reflect.Slice, reflect.Ptr, reflect.Chan, reflect.Func:
		return !v.IsNil()
	case reflect.UnsafePointer:
		return v.Pointer() != 0
	}
	return true
}

func CanExpand(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Map, reflect.Struct,
		reflect.Interface, reflect.Array, reflect.Slice,
		reflect.Ptr:
		return true
	}
	return false
}

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

func IsLabel(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Interface, reflect.Struct:
		return true
	}
	return false
}

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
