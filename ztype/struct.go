package ztype

import (
	"reflect"
)

type StructEngin struct {
	Fields        []interface{}
	Result        []map[string]interface{}
	TagName       string
	TagIgnoreName string
	ExtraCols     []string
}

func Struct() *StructEngin {
	s := new(StructEngin)
	s.TagName = "z"
	s.TagIgnoreName = "ignore"
	return s
}

func (s *StructEngin) GetStructFields(data interface{}) []interface{} {
	val := reflect.Indirect(reflect.ValueOf(data))
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		switch valueField.Kind() {
		case reflect.Struct:
			s.GetStructFields(valueField.Addr().Interface())
		default:
			s.AppendFields(valueField.Addr().Interface())
		}
	}
	return s.GetFields()
}

func (s *StructEngin) ToMap(data interface{}) []map[string]interface{} {
	val := reflect.Indirect(reflect.ValueOf(data))
	switch val.Kind() {
	case reflect.Struct:
		s.get(val)
	case reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			s.get(reflect.Indirect(val.Index(i)))
		}
	}
	return s.GetResult()
}

func (s *StructEngin) get(val reflect.Value) {
	valType := val.Type()
	var mapTmp = make(map[string]interface{})
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := valType.Field(i)
		switch valueField.Kind() {
		case reflect.Struct:
			continue
		default:
			fieldName := typeField.Tag.Get(s.GetTagName())
			if fieldName != s.GetTagIgnoreName() {
				if fieldName == "" {
					fieldName = typeField.Name
				}
				if ToBool(valueField.Interface()) {
					mapTmp[fieldName] = valueField.Interface()
				} else {
					if InArray(fieldName, s.ExtraCols) {
						mapTmp[fieldName] = valueField.Interface()
					}
				}
			}
		}
	}
	s.AppendResult(mapTmp)
}

func (s *StructEngin) AppendFields(arg interface{}) {
	s.Fields = append(s.Fields, arg)
}

func (s *StructEngin) SetFields(arg []interface{}) {
	s.Fields = arg
}
func (s *StructEngin) SetExtraCols(args []string) *StructEngin {
	s.ExtraCols = args
	return s
}

func (s *StructEngin) GetFields() []interface{} {
	return s.Fields
}

func (s *StructEngin) AppendResult(arg map[string]interface{}) {
	s.Result = append(s.Result, arg)
}

func (s *StructEngin) SetResult(arg []map[string]interface{}) {
	s.Result = arg
}

func (s *StructEngin) GetResult() []map[string]interface{} {
	return s.Result
}

func (s *StructEngin) SetTagName(arg string) *StructEngin {
	s.TagName = arg
	return s
}

func (s *StructEngin) GetTagName() string {
	return s.TagName
}

func (s *StructEngin) SetTagIgnoreName(arg string) *StructEngin {
	s.TagIgnoreName = arg
	return s
}

func (s *StructEngin) GetTagIgnoreName() string {
	return s.TagIgnoreName
}
