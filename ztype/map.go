package ztype

import (
	"errors"
	"reflect"
	"strings"
	"unsafe"

	"github.com/sohaha/zlsgo/zreflect"
	"github.com/sohaha/zlsgo/zstring"
)

var (
	tagName       = "z"
	tagNameLesser = "json"
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

func (m Map) Get(key string, disabled ...bool) Type {
	typ := Type{}
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

func (m Map) Set(key string, value interface{}) error {
	if m == nil {
		return errors.New("map is nil")
	}

	m[key] = value

	return nil
}

func (m Map) Has(key string) bool {
	_, ok := m[key]

	return ok
}

func (m Map) Delete(key string) error {
	if _, ok := m[key]; ok {
		delete(m, key)
		return nil
	}

	return errors.New("key not found")
}

func (m Map) ForEach(fn func(k string, v Type) bool) {
	for s, v := range m {
		if !fn(s, Type{v}) {
			return
		}
	}
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
		return toMapString(v)
	}
}

// ToMaps to Slice Map
func ToMaps(value interface{}) Maps {
	switch r := value.(type) {
	case Maps:
		return r
	case []map[string]interface{}:
		return *(*Maps)(unsafe.Pointer(&r))
	default:
		ref := reflect.Indirect(zreflect.ValueOf(value))
		m := make(Maps, 0)
		l := ref.Len()
		v := ref.Slice(0, l)
		for i := 0; i < l; i++ {
			m = append(m, toMapString(v.Index(i).Interface()))
		}
		return m
	}
}

func toMapString(value interface{}, tags ...string) map[string]interface{} {
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
		rv := zreflect.ValueOf(val)
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
		ol:
			for i := 0; i < rv.NumField(); i++ {
				field := rt.Field(i)
				fieldName := field.Name
				if !zstring.IsUcfirst(fieldName) {
					continue
				}

				name, opt := zreflect.GetStructTag(field, tagName, tagNameLesser)
				if name == "" {
					continue
				}
				array := strings.Split(opt, ",")
				v := rv.Field(i)
				for i := range array {
					switch strings.TrimSpace(array[i]) {
					case "omitempty":
						if IsEmpty(v.Interface()) {
							continue ol
						}
					}
				}
				fv := reflect.Indirect(v)
				switch fv.Kind() {
				case reflect.Struct:
					m[name] = toMapString(v.Interface())
					continue
				case reflect.Slice:
					if field.Type.Elem().Kind() == reflect.Struct {
						mc := make([]map[string]interface{}, v.Len())
						for i := 0; i < v.Len(); i++ {
							mc[i] = toMapString(v.Index(i).Interface())
						}
						m[name] = mc
						continue
					}
				}
				m[name] = v.Interface()
			}
		default:
			m["0"] = val
		}
	}
	return m

}

func ToMap2(v interface{}) Map {
	var m map[string]interface{}
	_ = conv.to("", v, zreflect.ValueOf(&m))
	return m
}
