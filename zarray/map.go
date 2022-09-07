//go:build go1.18
// +build go1.18

package zarray

// Keys creates an array of the map keys
func Keys[K comparable, V any](in map[K]V) []K {
	result := make([]K, 0, len(in))

	for k := range in {
		result = append(result, k)
	}

	return result
}

// Values creates an array of the map values
func Values[K comparable, V any](in map[K]V) []V {
	result := make([]V, 0, len(in))

	for _, v := range in {
		result = append(result, v)
	}

	return result
}
