package ztype

import (
	"errors"
	"reflect"
	"strings"

	"github.com/sohaha/zlsgo/zstring"
)

var (
	structTagPriority = []string{"zto", "json"}
)

type Map map[string]interface{}

func (m Map) DeepCopy() Map {
	newMap := make(Map, len(m))
	for k := range m {
		switch v := m[k].(type) {
		case Map:
			newMap[k] = v.DeepCopy()
		case map[string]interface{}:
			newMap[k] = Map(v).DeepCopy()
		default:
			newMap[k] = v
		}
	}

	return newMap
}

func (m Map) Get(key string, disabled ...bool) *Type {
	typ := &Type{}
	var (
		v  interface{}
		ok bool
	)
	if len(disabled) > 0 && disabled[0] {
		v, ok = m[key]
	} else {
		v, ok = parsePath(key, m)
	}
	if ok {
		typ.v = v
	}
	return typ
}

func (m *Map) Set(key string, value interface{}) error {
	if *m == nil {
		*m = make(Map)
	}
	(*m)[key] = value

	return nil
}

func (m Map) Has(key string) bool {
	_, ok := m[key]

	return ok
}

func (m *Map) Delete(key string) error {
	if _, ok := (*m)[key]; ok {
		delete(*m, key)
		return nil
	}
	return errors.New("key not found")
}

func (m Map) IsEmpty() bool {
	return len(m) == 0
}

type Maps []Map

func (m Maps) IsEmpty() bool {
	return len(m) == 0
}

func (m Maps) Len() int {
	return len(m)
}

func (m Maps) Index(i int) Map {
	if i < 0 || i >= len(m) {
		return Map{}
	}
	return m[i]
}

func (m Maps) ForEach(fn func(i int, value Map) bool) {
	for i := range m {
		v := m[i]
		if !fn(i, v) {
			break
		}
	}
}

// MapKeyExists Whether the dictionary key exists
func MapKeyExists(key interface{}, m map[interface{}]interface{}) bool {
	_, ok := m[key]
	return ok
}

func ToMap(value interface{}) Map {
	switch v := value.(type) {
	case Map:
		return v
	case map[string]interface{}:
		return v
	default:
		return ToMapString(v)
	}
}

// ToSliceMapString to SliceMapString
func ToSliceMapString(value interface{}) []map[string]interface{} {
	if r, ok := value.([]map[string]interface{}); ok {
		return r
	}
	ref := reflect.Indirect(reflect.ValueOf(value))
	m := make([]map[string]interface{}, 0)
	l := ref.Len()
	v := ref.Slice(0, l)
	for i := 0; i < l; i++ {
		m = append(m, ToMapString(v.Index(i).Interface()))
	}
	return m
}

// ToMapString ToMapString
func ToMapString(value interface{}, tags ...string) map[string]interface{} {
	if value == nil {
		return map[string]interface{}{}
	}
	if r, ok := value.(map[string]interface{}); ok {
		return r
	}
	m := map[string]interface{}{}
	switch val := value.(type) {
	case map[interface{}]interface{}:
		for k, v := range val {
			m[ToString(k)] = v
		}
	case map[interface{}]string:
		for k, v := range val {
			m[ToString(k)] = v
		}
	case map[interface{}]int:
		for k, v := range val {
			m[ToString(k)] = v
		}
	case map[interface{}]uint:
		for k, v := range val {
			m[ToString(k)] = v
		}
	case map[interface{}]float32:
		for k, v := range val {
			m[ToString(k)] = v
		}
	case map[interface{}]float64:
		for k, v := range val {
			m[ToString(k)] = v
		}
	case map[string]bool:
		for k, v := range val {
			m[k] = v
		}
	case map[string]int:
		for k, v := range val {
			m[k] = v
		}
	case map[string]uint:
		for k, v := range val {
			m[k] = v
		}
	case map[string]float32:
		for k, v := range val {
			m[k] = v
		}
	case map[string]float64:
		for k, v := range val {
			m[k] = v
		}
	case map[int]interface{}:
		for k, v := range val {
			m[ToString(k)] = v
		}
	case map[int]string:
		for k, v := range val {
			m[ToString(k)] = v
		}
	case map[uint]string:
		for k, v := range val {
			m[ToString(k)] = v
		}
	default:
		rv := reflect.ValueOf(val)
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
			m["0"] = val
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
