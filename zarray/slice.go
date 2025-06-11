//go:build go1.18
// +build go1.18

package zarray

import (
	"errors"
	"math/rand"
	"sort"

	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zsync"
	"github.com/sohaha/zlsgo/zutil"
)

// CopySlice creates and returns a new slice containing all elements from the input slice.
func CopySlice[T any](l []T) []T {
	nl := make([]T, len(l))
	copy(nl, l)
	return nl
}

// Rand returns a random element from the provided slice.
// If the slice is empty, returns the zero value of type T.
func Rand[T any](collection []T) T {
	l := len(collection)
	if l == 0 {
		var zero T
		return zero
	}

	i := zstring.RandInt(0, l-1)
	return collection[i]
}

// RandPickN returns a new slice containing n randomly selected elements from the input collection.
// If n is greater than the collection length, returns all elements in random order.
// If n is less than or equal to 0 or the collection is empty, returns an empty slice.
func RandPickN[T any](collection []T, n int) []T {
	l := len(collection)
	if l == 0 || n <= 0 {
		return []T{}
	}

	if n > l {
		n = l
	}

	temp := make([]T, l)
	copy(temp, collection)
	result := make([]T, n)

	for i := 0; i < n; i++ {
		j := zstring.RandInt(i, l-1)
		temp[i], temp[j] = temp[j], temp[i]
		result[i] = temp[i]
	}

	return result
}

// Map applies the iteratee function to each element in the collection and returns a new slice
// containing the transformed values. If parallel is provided, the operation is performed
// concurrently using the specified number of workers.
func Map[T any, R any](collection []T, iteratee func(int, T) R, parallel ...uint) []R {
	colLen := len(collection)
	res := make([]R, colLen)

	if len(parallel) == 0 {
		for i := range collection {
			res[i] = iteratee(i, collection[i])
		}
		return res
	}

	var (
		idx = zutil.NewInt64(0)
		wg  zsync.WaitGroup
	)

	task := func() {
		i := int(idx.Add(1) - 1)
		for ; i < colLen; i = int(idx.Add(1) - 1) {
			res[i] = iteratee(i, collection[i])
		}
	}

	workers := int(parallel[0])
	if workers > colLen || workers == 0 {
		workers = colLen
	}

	for i := 0; i < workers; i++ {
		wg.Go(task)
	}

	wg.Wait()

	return res
}

// ParallelMap applies the iteratee function to each element in the collection concurrently
// and returns a new slice containing the transformed values.
// If the calculation does not involve time-consuming operations, using Map is recommended.
// Deprecated: please use Map with a parallel parameter instead
func ParallelMap[T any, R any](collection []T, iteratee func(int, T) R, workers uint) []R {
	return Map(collection, iteratee, workers)
}

// Shuffle creates and returns a new slice containing all elements from the input slice
// in a random order. The original slice remains unchanged.
func Shuffle[T any](collection []T) []T {
	n := CopySlice(collection)
	rand.Shuffle(len(n), func(i, j int) {
		n[i], n[j] = n[j], n[i]
	})

	return n
}

// Reverse creates and returns a new slice containing all elements from the input slice
// in reverse order. The original slice remains unchanged.
func Reverse[T any](collection []T) []T {
	n := CopySlice(collection)
	l := len(n)
	for i := 0; i < l/2; i++ {
		n[i], n[l-i-1] = n[l-i-1], n[i]
	}

	return n
}

// Filter creates a new slice containing all elements from the input slice that satisfy
// the predicate function. The original slice remains unchanged.
func Filter[T any](slice []T, predicate func(index int, item T) bool) []T {
	slice = CopySlice(slice)

	j := 0
	for i := range slice {
		if !predicate(i, slice[i]) {
			continue
		}
		slice[j] = slice[i]
		j++
	}

	return slice[:j:j]
}

// Contains checks if a value exists in the collection.
// Returns true if the value is found, false otherwise.
func Contains[T comparable](collection []T, v T) bool {
	for i := range collection {
		if collection[i] == v {
			return true
		}
	}

	return false
}

// Find searches for an element in the slice that satisfies the predicate function.
// Returns the found element and true if successful, or the zero value and false if not found.
func Find[T any](collection []T, predicate func(index int, item T) bool) (res T, ok bool) {
	for i := range collection {
		item := collection[i]
		if predicate(i, item) {
			return item, true
		}
	}

	return
}

// Unique creates and returns a new slice containing only the unique elements
// from the input slice, preserving the original order of first occurrence.
func Unique[T comparable](collection []T) []T {
	l := len(collection)
	if l <= 1 {
		return CopySlice(collection)
	}

	repeat := make(map[T]struct{}, len(collection))

	return Filter(collection, func(_ int, item T) bool {
		if _, ok := repeat[item]; ok {
			return false
		}
		repeat[item] = struct{}{}
		return true
	})
}

// Diff compares two slices and returns two new slices containing the elements that
// are unique to each input slice (not present in the other slice).
func Diff[T comparable](list1 []T, list2 []T) ([]T, []T) {
	if len(list1) == 0 {
		return []T{}, CopySlice(list2)
	}
	if len(list2) == 0 {
		return CopySlice(list1), []T{}
	}

	rl := make(map[T]struct{}, len(list1))
	rr := make(map[T]struct{}, len(list2))

	for _, e := range list1 {
		rl[e] = struct{}{}
	}

	for _, e := range list2 {
		rr[e] = struct{}{}
	}

	l := make([]T, 0, len(list1))
	r := make([]T, 0, len(list2))

	for _, e := range list1 {
		if _, ok := rr[e]; !ok {
			l = append(l, e)
		}
	}

	for _, e := range list2 {
		if _, ok := rl[e]; !ok {
			r = append(r, e)
		}
	}

	return l, r
}

// Pop removes and returns the last element from the slice.
// If the slice is empty, returns the zero value of type T.
// This function modifies the original slice.
func Pop[T comparable](list *[]T) (v T) {
	l := len(*list)
	if l == 0 {
		return
	}

	v = (*list)[l-1]
	*list = (*list)[:l-1]
	return
}

// Shift removes and returns the first element from the slice.
// If the slice is empty, returns the zero value of type T.
// This function modifies the original slice.
func Shift[T comparable](list *[]T) (v T) {
	l := len(*list)
	if l == 0 {
		return
	}

	v = (*list)[0]
	*list = (*list)[1:]
	return
}

// Chunk splits the slice into multiple sub-slices of the specified size.
// The last chunk may contain fewer elements if the slice length is not divisible by size.
// If size is less than or equal to 0 or the slice is empty, returns an empty slice of slices.
func Chunk[T any](slice []T, size int) [][]T {
	if size <= 0 {
		return [][]T{}
	}

	l := len(slice)
	if l == 0 {
		return [][]T{}
	}

	n := (l + size - 1) / size
	chunks := make([][]T, n)

	for i := 0; i < n-1; i++ {
		chunks[i] = slice[i*size : (i+1)*size]
	}

	chunks[n-1] = slice[(n-1)*size:]
	return chunks
}

// RandShift returns a closure function that, when called, returns a random element from the list.
// Each element is returned exactly once in random order. When all elements have been returned,
// subsequent calls will return an error. The original list is not modified.
func RandShift[T comparable](list []T) func() (T, error) {
	if len(list) == 0 {
		return func() (T, error) {
			var zero T
			return zero, errors.New("no available items to select randomly")
		}
	}

	indices := rand.Perm(len(list))
	currentIndex := 0

	return func() (T, error) {
		if currentIndex >= len(indices) {
			var zero T
			return zero, errors.New("no available items to select randomly")
		}

		item := list[indices[currentIndex]]
		currentIndex++
		return item, nil
	}
}

// SortWithPriority sorts the slice based on the priority of elements.
// The elements in the 'first' slice are placed at the beginning of the result,
// followed by the elements in the 'last' slice, and then the remaining elements.
func SortWithPriority[T comparable](slice []T, first, last []T) []T {
	if len(slice) == 0 {
		return []T{}
	}

	result := make([]T, len(slice))
	copy(result, slice)

	ranks := make(map[T]int, len(first)+len(last))
	for i, h := range first {
		ranks[h] = i
	}

	lastRankStart := len(first) + len(slice)
	for i, f := range last {
		ranks[f] = lastRankStart + i
	}

	defaultRank := len(first)
	sort.SliceStable(result, func(i, j int) bool {
		rankI, okI := ranks[result[i]]
		if !okI {
			rankI = defaultRank
		}

		rankJ, okJ := ranks[result[j]]
		if !okJ {
			rankJ = defaultRank
		}

		return rankI < rankJ
	})

	return result
}
