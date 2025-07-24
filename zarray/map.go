//go:build go1.18
// +build go1.18

package zarray

import "errors"

// Keys extracts all keys from a map and returns them as a slice.
// The order of the keys in the resulting slice is not guaranteed.
func Keys[K comparable, V any](in map[K]V) []K {
	result := make([]K, 0, len(in))

	for k := range in {
		result = append(result, k)
	}

	return result
}

// Values extracts all values from a map and returns them as a slice.
// The order of the values in the resulting slice is not guaranteed.
func Values[K comparable, V any](in map[K]V) []V {
	result := make([]V, 0, len(in))

	for _, v := range in {
		result = append(result, v)
	}

	return result
}

// IndexMap indexes a slice of Maps into a map based on a key function.
func IndexMap[K comparable, V any](arr []V, toKey func(V) (K, V)) (map[K]V, error) {
	if len(arr) == 0 {
		return make(map[K]V), nil
	}

	data := make(map[K]V, len(arr))
	for _, item := range arr {
		key, value := toKey(item)
		if _, exists := data[key]; exists {
			return nil, errors.New("key is not unique")
		}
		data[key] = value
	}
	return data, nil
}

// FlatMap flattens a map of Maps into a single slice of Maps.
func FlatMap[K comparable, V any](m map[K]V, fn func(key K, value V) V) []V {
	data := make([]V, 0, len(m))
	for k := range m {
		data = append(data, fn(k, m[k]))
	}
	return data
}
