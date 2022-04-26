package ztype

import "reflect"

func Slice(value interface{}) (m []interface{}) {
	ref := reflect.Indirect(reflect.ValueOf(value))
	switch ref.Kind() {
	case reflect.Slice, reflect.String:
		l := ref.Len()
		v := ref.Slice(0, l)
		for i := 0; i < l; i++ {
			m = append(m, v.Index(i).Interface())
		}
	}
	return m
}

// SliceStrToIface  []string to []interface{}
func SliceStrToIface(slice []string) []interface{} {
	ifeSlice := make([]interface{}, 0, len(slice))
	for _, val := range slice {
		ifeSlice = append(ifeSlice, val)
	}
	return ifeSlice
}
