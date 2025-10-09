package ztype

import (
	"reflect"
	"strings"
	"sync"

	"github.com/sohaha/zlsgo/zreflect"
	"github.com/sohaha/zlsgo/zstring"
)

// fieldInfo cache struct field info
type fieldInfo struct {
	Type       reflect.Type
	Name       string
	Tag        string
	Options    []string
	Index      int
	IsTime     bool
	IsExported bool
}

// structCacheEntry struct cache entry
type structCacheEntry struct {
	TypeName string
	Fields   []fieldInfo
}

// structCache struct cache using reflect.Type as key to avoid name conflicts
var structCache sync.Map

// getStructInfo get struct info, priority from cache
func getStructInfo(t reflect.Type) []fieldInfo {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	if cached, exists := structCache.Load(t); exists {
		return cached.(structCacheEntry).Fields
	}

	fields := parseStructFields(t)
	entry := structCacheEntry{
		Fields:   fields,
		TypeName: t.String(),
	}

	if actual, loaded := structCache.LoadOrStore(t, entry); loaded {
		return actual.(structCacheEntry).Fields
	}

	return fields
}

// parseStructFields parse struct fields
func parseStructFields(t reflect.Type) []fieldInfo {
	var fields []fieldInfo

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldName := field.Name

		if !zstring.IsUcfirst(fieldName) {
			continue
		}

		name, opt := zreflect.GetStructTag(field, tagName, tagNameLesser)
		if name == "" {
			continue
		}

		options := parseTagOptions(opt)

		isTimeType := isTime(field.Type.String())

		fieldInfo := fieldInfo{
			Name:       name,
			Index:      i,
			Tag:        string(field.Tag),
			Type:       field.Type,
			IsTime:     isTimeType,
			Options:    options,
			IsExported: true,
		}

		fields = append(fields, fieldInfo)
	}

	return fields
}

// parseTagOptions parse tag options
func parseTagOptions(opt string) []string {
	if opt == "" {
		return nil
	}

	options := getStringSlice()
	parts := strings.Split(opt, ",")
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			options = append(options, trimmed)
		}
	}

	if len(options) == 0 {
		putStringSlice(options)
		return nil
	}
	result := make([]string, len(options))
	copy(result, options)
	putStringSlice(options)

	return result
}

// hasOption check field has specific option
func (f *fieldInfo) hasOption(option string) bool {
	for _, opt := range f.Options {
		if opt == option {
			return true
		}
	}
	return false
}
