//go:build go1.18
// +build go1.18

package zarray

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
