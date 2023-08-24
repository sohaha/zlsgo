package ztype

import (
	"encoding/json"

	"github.com/sohaha/zlsgo/zreflect"
)

type SliceType []Type

func (s SliceType) Len() int {
	return len(s)
}

func (s SliceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Value())
}

func (s SliceType) Index(i int) Type {
	if len(s) <= i {
		return Type{}
	}
	return s[i]
}

func (s SliceType) Value() []interface{} {
	ss := make([]interface{}, 0, len(s))
	for i := range s {
		ss = append(ss, s[i].Value())
	}
	return ss
}

func (s SliceType) String() []string {
	ss := make([]string, 0, len(s))
	for i := range s {
		ss = append(ss, s[i].String())
	}
	return ss
}

func (s SliceType) Int() []int {
	ss := make([]int, 0, len(s))
	for i := range s {
		ss = append(ss, s[i].Int())
	}
	return ss
}

func (s SliceType) Maps() Maps {
	ss := make(Maps, 0, len(s))
	for i := range s {
		ss = append(ss, s[i].Map())
	}
	return ss
}

// Deprecated: please use ToSlice
func Slice(value interface{}) SliceType {
	return ToSlice(value)
}

// SliceStrToAny  []string to []interface{}
func SliceStrToAny(slice []string) []interface{} {
	ss := make([]interface{}, 0, len(slice))
	for _, val := range slice {
		ss = append(ss, val)
	}
	return ss
}

func ToSlice(value interface{}) SliceType {
	s := SliceType{}
	if value == nil {
		return s
	}

	switch val := value.(type) {
	case []interface{}:
		s = make(SliceType, 0, len(val))
		for i := range val {
			s = append(s, New(val[i]))
		}
	case []string:
		s = make(SliceType, 0, len(val))
		for i := range val {
			s = append(s, New(val[i]))
		}
	case string:
		if val != "" {
			s = append(s, New(val))
		}
	default:
		var nval []interface{}
		if conv.to("", value, zreflect.ValueOf(&nval)) == nil {
			s = make(SliceType, 0, len(nval))
			for i := range nval {
				s = append(s, New(nval[i]))
			}
		}
	}

	return s
}
