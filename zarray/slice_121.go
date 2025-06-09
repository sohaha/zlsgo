//go:build go1.21
// +build go1.21

package zarray

// Intersection returns a new slice containing the common elements from two slices.
// The order of elements in the returned slice is based on the order in list1.
// Duplicate elements in the intersection are removed.
func Intersection[T comparable](list1 []T, list2 []T) []T {
	if len(list1) == 0 || len(list2) == 0 {
		return []T{}
	}

	set2 := make(map[T]struct{}, len(list2))
	for _, item := range list2 {
		set2[item] = struct{}{}
	}

	result := make([]T, 0, min(len(list1), len(list2)))
	seenInResult := make(map[T]struct{}, min(len(list1), len(list2)))

	for _, item := range list1 {
		if _, ok := set2[item]; ok {
			if _, seen := seenInResult[item]; !seen {
				result = append(result, item)
				seenInResult[item] = struct{}{}
			}
		}
	}

	return result
}
