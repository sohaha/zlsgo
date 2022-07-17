package ztype

import (
	"reflect"
)

// Deprecated: please use ToSlice
func Slice(value interface{}) []interface{} {
	return ToSlice(value)
}

// SliceStrToIface  []string to []interface{}
func SliceStrToIface(slice []string) []interface{} {
	ifeSlice := make([]interface{}, 0, len(slice))
	for _, val := range slice {
		ifeSlice = append(ifeSlice, val)
	}
	return ifeSlice
}

func ToSlice(value interface{}) (s []interface{}) {
	ref := reflect.Indirect(reflect.ValueOf(value))
	switch ref.Kind() {
	case reflect.Slice:
		l := ref.Len()
		v := ref.Slice(0, l)
		for i := 0; i < l; i++ {
			s = append(s, v.Index(i).Interface())
		}
	default:
		s = append(s, value)
	}
	return s
}
