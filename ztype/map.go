package ztype

import (
	"errors"
	"reflect"
	"time"
	"unsafe"

	"github.com/sohaha/zlsgo/zreflect"
	"github.com/sohaha/zlsgo/ztime"
)

var (
	// tagName is the primary struct tag name used for field mapping
	tagName = "z"
	// tagNameLesser is the fallback struct tag name used when the primary tag is not present
	tagNameLesser = "json"
)

// Map is a string-keyed map of arbitrary values that provides helper methods
// for convenient access and manipulation of the underlying data.
type Map map[string]interface{}

// DeepCopy creates a deep copy of the map and all nested maps.
// This ensures that modifications to the copied map don't affect the original.
func (m Map) DeepCopy() Map {
	newMap := make(map[string]interface{})
	for k := range m {
		switch v := m[k].(type) {
		case Map:
			if v == nil {
				newMap[k] = v
				continue
			}
			newMap[k] = v.DeepCopy()
		case map[string]interface{}:
			newMap[k] = Map(v).DeepCopy()
		default:
			typ := zreflect.TypeOf(v)
			if typ.Kind() == reflect.Map {
				newMap[k] = ToMap(v).DeepCopy()
			} else {
				newMap[k] = v
			}
		}
	}

	return newMap
}

// Get retrieves a value from the map by its key and wraps it in a Type for safe access.
// If disabled is true, it will only look for exact key matches and not parse path expressions.
// Path expressions (like "user.name" or "items[0].id") allow accessing nested values.
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

// Set assigns a value to the specified key in the map.
// Returns an error if the map is nil.
func (m Map) Set(key string, value interface{}) error {
	if m == nil {
		return errors.New("map is nil")
	}

	m[key] = value

	return nil
}

// Has checks if the specified key exists in the map.
// Returns true if the key exists, false otherwise.
func (m Map) Has(key string) bool {
	_, ok := m[key]

	return ok
}

// Delete removes a key-value pair from the map.
// Returns an error if the key doesn't exist.
func (m Map) Delete(key string) error {
	if _, ok := m[key]; ok {
		delete(m, key)
		return nil
	}

	return errors.New("key not found")
}

// Valid checks if the specified keys exist in the map.
// Returns true if all keys exist, false otherwise.
func (m Map) Valid(keys ...string) bool {
	if m == nil {
		return false
	}
	for _, key := range keys {
		if _, ok := m[key]; !ok {
			return false
		}
	}
	return true
}

// Keys returns a slice containing all keys currently in the map.
func (m Map) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// ForEach iterates over all key-value pairs in the map and calls the provided function for each pair.
// If the function returns false, iteration stops.
func (m Map) ForEach(fn func(k string, v Type) bool) {
	for s, v := range m {
		if !fn(s, Type{v}) {
			return
		}
	}
}

// IsEmpty checks if the map contains any elements.
// Returns true if the map is empty, false otherwise.
func (m Map) IsEmpty() bool {
	return len(m) == 0
}

// Maps is a slice of Map objects, providing helper methods for working with collections of maps.
type Maps []Map

// IsEmpty checks if the slice contains any maps.
// Returns true if the slice is empty, false otherwise.
func (m Maps) IsEmpty() bool {
	return len(m) == 0
}

// Len returns the number of maps in the slice.
func (m Maps) Len() int {
	return len(m)
}

// Index returns the map at the specified index.
// Returns an empty map if the index is out of bounds.
func (m Maps) Index(i int) Map {
	if i < 0 || i >= len(m) {
		return Map{}
	}
	return m[i]
}

// Last returns the last map in the slice.
// Returns an empty map if the slice is empty.
func (m Maps) Last() Map {
	l := m.Len()
	if l == 0 {
		return Map{}
	}
	return m[l-1]
}

// First returns the first map in the slice.
// Returns an empty map if the slice is empty.
func (m Maps) First() Map {
	return m.Index(0)
}

// ForEach iterates over all maps in the slice and calls the provided function for each one.
// If the function returns false, iteration stops.
func (m Maps) ForEach(fn func(i int, value Map) bool) {
	for i := range m {
		v := m[i]
		if !fn(i, v) {
			break
		}
	}
}

// MapKeyExists checks if a key exists in a map with interface{} keys and values.
// Returns true if the key exists, false otherwise.
func MapKeyExists(key interface{}, m map[interface{}]interface{}) bool {
	_, ok := m[key]
	return ok
}

// ToMap converts various types to a Map.
// Handles Map, map[string]interface{}, and struct types through reflection.
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

// ToMaps converts various types to a Maps slice.
// Handles Maps, []map[string]interface{}, and slices of structs through reflection.
func ToMaps(value interface{}) Maps {
	switch r := value.(type) {
	case Maps:
		return r
	case []map[string]interface{}:
		return *(*Maps)(unsafe.Pointer(&r))
	default:
		ref := reflect.Indirect(zreflect.ValueOf(value))
		l := ref.Len()
		v := ref.Slice(0, l)

		result := make(Maps, 0, l)
		for i := 0; i < l; i++ {
			result = append(result, toMapString(v.Index(i).Interface()))
		}
		return result
	}
}

// toMapString converts various map types to a map[string]interface{}.
// This is an internal function used by ToMap to handle different map types.
func toMapString(value interface{}) map[string]interface{} {
	if value == nil {
		return make(map[string]interface{}, 0)
	}
	if r, ok := value.(map[string]interface{}); ok {
		return r
	}

	var capacity int
	switch v := value.(type) {
	case map[interface{}]interface{}:
		capacity = len(v)
	case map[interface{}]string:
		capacity = len(v)
	case map[string]bool:
		capacity = len(v)
	case map[string]int:
		capacity = len(v)
	default:
		capacity = 8
	}

	m := make(map[string]interface{}, capacity)
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
		toMapStringReflect(&m, value)
	}
	return m
}

func toMapStringReflect(m *map[string]interface{}, val interface{}) {
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
			(*m)[ToString(k.Interface())] = rv.MapIndex(k).Interface()
		}
	case reflect.Struct:
		fields := getStructInfo(rv.Type())
	ol:
		for _, fieldInfo := range fields {
			v := rv.Field(fieldInfo.Index)

			if fieldInfo.hasOption("omitempty") {
				if IsEmpty(v.Interface()) {
					continue ol
				}
			}

			fv := reflect.Indirect(v)

			if fv.IsValid() && fieldInfo.IsTime {
				switch val := fv.Interface().(type) {
				case time.Time:
					(*m)[fieldInfo.Name] = ztime.FormatTime(val)
				case ztime.LocalTime:
					(*m)[fieldInfo.Name] = val.String()
				}
				continue
			}

			switch fv.Kind() {
			case reflect.Struct:
				(*m)[fieldInfo.Name] = toMapString(v.Interface())
				continue
			case reflect.Slice:
				if fieldInfo.Type.Elem().Kind() == reflect.Struct {
					mc := getMapSlice()
					for i := 0; i < v.Len(); i++ {
						mc = append(mc, toMapString(v.Index(i).Interface()))
					}
					result := make([]map[string]interface{}, len(mc))
					copy(result, mc)
					putMapSlice(mc)
					(*m)[fieldInfo.Name] = result
					continue
				}
			}
			(*m)[fieldInfo.Name] = v.Interface()
		}
	default:
		(*m)["0"] = val
	}
}

// ValidateOptions configures validation behavior
type ValidateOptions struct {
	FastPath            bool // Use optimized fast path for small datasets
	UnsafeMode          bool // Skip safety checks for maximum performance
	ConcurrentThreshold int  // Threshold for concurrent validation (default: 500)
}

// Validate validates a single field with the given validator
func (m Map) Validate(key string, validator Validator) error {
	if m == nil {
		return ErrMapNil
	}
	if validator == nil {
		return ErrValidatorNil
	}

	value, exists := m[key]
	if !exists {
		return newKeyNotFoundError(key)
	}

	result := validator.VerifyAny(value, key)
	if result == nil {
		return ErrNilResult
	}

	return result.Error()
}

// ValidateAll validates all fields according to the provided rules
func (m Map) ValidateAll(rules map[string]Validator) error {
	if m == nil {
		return ErrMapNil
	}
	if len(rules) == 0 {
		return nil
	}

	for key, validator := range rules {
		if err := m.Validate(key, validator); err != nil {
			return err
		}
	}

	return nil
}

// ValidateWithOptions validates all fields with configurable options
func (m Map) ValidateWithOptions(rules map[string]Validator, opts ...ValidateOptions) error {
	if m == nil {
		return ErrMapNil
	}
	if len(rules) == 0 {
		return nil
	}

	// Use default options if none provided
	if len(opts) > 0 {
		_ = opts[0] // Options acknowledged but not used in simplified implementation
	}

	// For simplicity, all options route to the same sequential validation
	// The intelligent selection logic was removed as per user request
	return m.ValidateAll(rules)
}

// Validator interface defines the validation contract
type Validator interface {
	VerifyAny(value interface{}, name ...string) ValidatorResult
}

// ValidatorResult interface defines the validation result contract
type ValidatorResult interface {
	Error() error
}

// Error constants
var (
	ErrMapNil       = errors.New("map is nil")
	ErrValidatorNil = errors.New("validator is nil")
	ErrNilResult    = errors.New("validator returned nil result")
)

// newKeyNotFoundError creates a simple error without pool management for safe usage
func newKeyNotFoundError(key string) error {
	return errors.New("key '" + key + "' not found")
}
