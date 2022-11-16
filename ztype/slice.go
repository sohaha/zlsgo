package ztype

import (
	"reflect"
)

type SliceType []Type

func (s SliceType) Len() int {
	return len(s)
}

func (s SliceType) Index(i int) Type {
	if len(s) <= i {
		return Type{}
	}
	return s[i]
}

func (s SliceType) Value() []interface{} {
	ss := make([]interface{}, 0, len(s))
	for _, val := range s {
		ss = append(ss, val.Value())
	}
	return ss
}

func (s SliceType) String() []string {
	ss := make([]string, 0, len(s))
	for _, val := range s {
		ss = append(ss, val.String())
	}
	return ss
}

func (s SliceType) Int() []int {
	ss := make([]int, 0, len(s))
	for _, val := range s {
		ss = append(ss, val.Int())
	}
	return ss
}

// Deprecated: please use ToSlice
func Slice(value interface{}) SliceType {
	return ToSlice(value)
}

// SliceStrToIface  []string to []interface{}
func SliceStrToIface(slice []string) []interface{} {
	ss := make([]interface{}, 0, len(slice))
	for _, val := range slice {
		ss = append(ss, val)
	}
	return ss
}

func ToSlice(value interface{}) (s SliceType) {
	s = make(SliceType, 0)
	if value == nil {
		return
	}

	switch val := value.(type) {
	case []interface{}:
		s = make(SliceType, 0, len(val))
		for _, v := range val {
			s = append(s, New(v))
		}
	case []string:
		s = make(SliceType, 0, len(val))
		for _, v := range val {
			s = append(s, New(v))

		}
	case string:
		if val != "" {
			s = append(s, New(val))
		}
	default:
		ref := reflect.Indirect(reflect.ValueOf(value))
		switch ref.Kind() {
		case reflect.Slice:
			l := ref.Len()
			v := ref.Slice(0, l)
			for i := 0; i < l; i++ {
				val := v.Index(i).Interface()
				s = append(s, New(val))
			}
		default:
			s = append(s, New(value))
		}
	}
	return s
}
