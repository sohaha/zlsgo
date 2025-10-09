package ztype

import (
	"encoding/json"
	"reflect"

	"github.com/sohaha/zlsgo/zreflect"
)

// SliceType is a slice of Type objects that provides helper methods for
// working with collections of values with automatic type conversion.
type SliceType []Type

// Len returns the number of elements in the slice.
func (s SliceType) Len() int {
	return len(s)
}

// MarshalJSON implements the json.Marshaler interface.
// It marshals the underlying values rather than the Type wrappers.
func (s SliceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Value())
}

// Index returns the element at the specified index.
// Returns an empty Type if the index is out of bounds.
func (s SliceType) Index(i int) Type {
	if i < 0 || i >= len(s) {
		return Type{}
	}
	return s[i]
}

// Last returns the last element in the slice.
// Returns an empty Type if the slice is empty.
func (s SliceType) Last() Type {
	return s.Index(len(s) - 1)
}

// First returns the first element in the slice.
// Returns an empty Type if the slice is empty.
func (s SliceType) First() Type {
	return s.Index(0)
}

// Value returns the underlying values as a slice of interface{}.
// This unwraps all Type objects to their original values.
func (s SliceType) Value() []interface{} {
	if len(s) == 0 {
		return []interface{}{}
	}

	ss := getInterfaceSlice()
	if cap(ss) < len(s) {
		ss = make([]interface{}, 0, len(s))
	}
	for i := range s {
		ss = append(ss, s[i].Value())
	}
	result := make([]interface{}, len(ss))
	copy(result, ss)
	putInterfaceSlice(ss)
	return result
}

// String converts all elements in the slice to strings and returns them as a []string.
// Each element is converted using the Type.String() method.
func (s SliceType) String() []string {
	if len(s) == 0 {
		return []string{}
	}

	ss := getStringSlice()
	if cap(ss) < len(s) {
		ss = make([]string, 0, len(s))
	}
	for i := range s {
		ss = append(ss, s[i].String())
	}
	result := make([]string, len(ss))
	copy(result, ss)
	putStringSlice(ss)
	return result
}

// Int converts all elements in the slice to integers and returns them as a []int.
// Each element is converted using the Type.Int() method.
func (s SliceType) Int() []int {
	if len(s) == 0 {
		return []int{}
	}

	ss := getIntSlice()
	if cap(ss) < len(s) {
		ss = make([]int, 0, len(s))
	}
	for i := range s {
		ss = append(ss, s[i].Int())
	}
	result := make([]int, len(ss))
	copy(result, ss)
	putIntSlice(ss)
	return result
}

// Maps converts all elements in the slice to Map objects and returns them as a Maps slice.
// Each element is converted using the Type.Map() method.
func (s SliceType) Maps() Maps {
	ss := make(Maps, 0, len(s))
	for i := range s {
		ss = append(ss, s[i].Map())
	}
	return ss
}

// func (s SliceType) Slice() []SliceType {
// 	ss := make([]SliceType, 0, len(s))
// 	for i := range s {
// 		ss = append(ss, s[i].Slice())
// 	}
// 	return ss
// }

// Slice converts a value to a SliceType.
// Deprecated: please use ToSlice instead.
func Slice(value interface{}, noConv ...bool) SliceType {
	return ToSlice(value, noConv...)
}

// SliceStrToAny converts a slice of strings to a slice of interface{} values.
// This is useful when you need to pass a string slice to a function that expects interface{} values.
func SliceStrToAny(slice []string) []interface{} {
	ss := make([]interface{}, 0, len(slice))
	for _, val := range slice {
		ss = append(ss, val)
	}
	return ss
}

// ToSlice converts various types to a SliceType.
// If noConv is true, it will not attempt to convert non-slice values (like strings) to slices.
// Handles []interface{}, []string, []int, []int64, and can parse JSON strings into slices.
func ToSlice(value interface{}, noConv ...bool) (s SliceType) {
	s = SliceType{}
	if value == nil {
		return
	}
	nc := len(noConv) > 0 && noConv[0]
	switch val := value.(type) {
	case []interface{}:
		s = make(SliceType, len(val))
		for i := range val {
			s[i] = New(val[i])
		}
	case []string:
		s = make(SliceType, len(val))
		for i := range val {
			s[i] = New(val[i])
		}
	case []int:
		s = make(SliceType, len(val))
		for i := range val {
			s[i] = New(val[i])
		}
	case []int64:
		s = make(SliceType, len(val))
		for i := range val {
			s[i] = New(val[i])
		}
	case string:
		if nc {
			return
		}
		var nval []interface{}
		if err := json.Unmarshal([]byte(val), &nval); err == nil {
			s = make(SliceType, len(nval))
			for i := range nval {
				s[i] = New(nval[i])
			}
			return
		}
		s = SliceType{New(val)}
	case Type:
		return ToSlice(val.Value())
	default:
		var nval []interface{}
		vof := zreflect.ValueOf(&nval)
		to := func() {
			if conv.to("", value, vof, true) == nil {
				s = make(SliceType, len(nval))
				for i := range nval {
					s[i] = New(nval[i])
				}
			}
		}

		switch vof.Type().Kind() {
		case reflect.Slice:
			to()
		default:
			if nc {
				return
			}
			to()
		}
	}

	return
}
