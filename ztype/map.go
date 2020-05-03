package ztype

import (
	"reflect"
	"strings"

	"github.com/sohaha/zlsgo/zstring"
)

var (
	structTagPriority = []string{"zto", "c", "json"}
)

// MapKeyExists Whether the dictionary key exists
func MapKeyExists(key interface{}, m map[interface{}]interface{}) bool {
	_, ok := m[key]
	return ok
}

// ToSliceMapString ToSliceMapString
func ToSliceMapString(value interface{}, tags ...string) []map[string]interface{} {
	if r, ok := value.([]map[string]interface{}); ok {
		return r
	}
	ref := reflect.Indirect(reflect.ValueOf(value))
	m := []map[string]interface{}{}
	// switch ref.Kind() {
	// case reflect.Slice:
	l := ref.Len()
	v := ref.Slice(0, l)
	for i := 0; i < l; i++ {
		m = append(m, ToMapString(v.Index(i).Interface()))
	}
	// }
	return m
}

// ToMapString ToMapString
func ToMapString(value interface{}, tags ...string) map[string]interface{} {
	if value == nil {
		return nil
	}
	if r, ok := value.(map[string]interface{}); ok {
		return r
	}
	m := map[string]interface{}{}
	switch value := value.(type) {
	case map[interface{}]interface{}:
		for k, v := range value {
			m[ToString(k)] = v
		}
	case map[interface{}]string:
		for k, v := range value {
			m[ToString(k)] = v
		}
	case map[interface{}]int:
		for k, v := range value {
			m[ToString(k)] = v
		}
	case map[interface{}]uint:
		for k, v := range value {
			m[ToString(k)] = v
		}
	case map[interface{}]float32:
		for k, v := range value {
			m[ToString(k)] = v
		}
	case map[interface{}]float64:
		for k, v := range value {
			m[ToString(k)] = v
		}
	case map[string]bool:
		for k, v := range value {
			m[k] = v
		}
	case map[string]int:
		for k, v := range value {
			m[k] = v
		}
	case map[string]uint:
		for k, v := range value {
			m[k] = v
		}
	case map[string]float32:
		for k, v := range value {
			m[k] = v
		}
	case map[string]float64:
		for k, v := range value {
			m[k] = v
		}
	case map[int]interface{}:
		for k, v := range value {
			m[ToString(k)] = v
		}
	case map[int]string:
		for k, v := range value {
			m[ToString(k)] = v
		}
	case map[uint]string:
		for k, v := range value {
			m[ToString(k)] = v
		}
	default:
		rv := reflect.ValueOf(value)
		kind := rv.Kind()
		if kind == reflect.Ptr {
			rv = rv.Elem()
			kind = rv.Kind()
		}
		switch kind {
		case reflect.Map:
			ks := rv.MapKeys()
			for _, k := range ks {
				m[ToString(k.Interface())] = rv.MapIndex(k).Interface()
			}
		case reflect.Struct:
			rt := rv.Type()
			name := ""
			tagArray := structTagPriority
			switch len(tags) {
			case 0:
			case 1:
				tagArray = append(strings.Split(tags[0], ","), structTagPriority...)
			default:
				tagArray = append(tags, structTagPriority...)
			}
			for i := 0; i < rv.NumField(); i++ {
				fieldName := rt.Field(i).Name
				if !zstring.IsUcfirst(fieldName) {
					continue
				}
				name = ""
				fieldTag := rt.Field(i).Tag
				for _, tag := range tagArray {
					if name = fieldTag.Get(tag); name != "" {
						break
					}
				}
				if name == "" {
					name = strings.TrimSpace(fieldName)
				} else {
					name = strings.TrimSpace(name)
					if name == "-" {
						continue
					}
					array := strings.Split(name, ",")
					if len(array) > 1 {
						switch strings.TrimSpace(array[1]) {
						case "omitempty":
							if IsEmpty(rv.Field(i).Interface()) {
								continue
							} else {
								name = strings.TrimSpace(array[0])
							}
						default:
							name = strings.TrimSpace(array[0])
						}
					}
				}
				m[name] = rv.Field(i).Interface()
			}
		default:
			return nil
		}
	}
	return m

}

func ToMapStringDeep(value interface{}, tags ...string) map[string]interface{} {
	data := ToMapString(value, tags...)
	for key, value := range data {
		rv := reflect.ValueOf(value)
		kind := rv.Kind()
		if kind == reflect.Ptr {
			rv = rv.Elem()
			kind = rv.Kind()
		}
		switch kind {
		case reflect.Struct:
			delete(data, key)
			for k, v := range ToMapStringDeep(value, tags...) {
				data[k] = v
			}
		}
	}
	return data
}
