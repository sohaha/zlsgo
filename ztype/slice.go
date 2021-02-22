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
