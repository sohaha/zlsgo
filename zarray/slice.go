//go:build go1.18
// +build go1.18

package zarray

import (
	"errors"
	"math/rand"

	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zsync"
	"github.com/sohaha/zlsgo/zutil"
)

// CopySlice copy a slice.
func CopySlice[T any](l []T) []T {
	nl := make([]T, len(l))
	copy(nl, l)
	return nl
}

// Rand A random eents.
func Rand[T any](collection []T) T {
	l := len(collection)
	if l == 0 {
		var zero T
		return zero
	}

	i := zstring.RandInt(0, l-1)
	return collection[i]
}

// RandPickN returns a random slice of n elements from the collection.
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

// Map manipulates a slice and transforms it to a slice of another type.
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

// ParallelMap Parallel manipulates a slice and transforms it to a slice of another type.
// If the calculation does not involve time-consuming operations, we recommend using a Map.
// Deprecated: please use Map
func ParallelMap[T any, R any](collection []T, iteratee func(int, T) R, workers uint) []R {
	return Map(collection, iteratee, workers)
}

// Shuffle creates a slice of shuffled values.
func Shuffle[T any](collection []T) []T {
	n := CopySlice(collection)
	rand.Shuffle(len(n), func(i, j int) {
		n[i], n[j] = n[j], n[i]
	})

	return n
}

// Reverse creates a slice of reversed values.
func Reverse[T any](collection []T) []T {
	n := CopySlice(collection)
	l := len(n)
	for i := 0; i < l/2; i++ {
		n[i], n[l-i-1] = n[l-i-1], n[i]
	}

	return n
}

// Filter iterates over eents of collection.
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

// Contains returns true if an eent is present in a collection.
func Contains[T comparable](collection []T, v T) bool {
	for i := range collection {
		if collection[i] == v {
			return true
		}
	}

	return false
}

// Find search an eent in a slice based on a predicate. It returns eent and true if eent was found.
func Find[T any](collection []T, predicate func(index int, item T) bool) (res T, ok bool) {
	for i := range collection {
		item := collection[i]
		if predicate(i, item) {
			return item, true
		}
	}

	return
}

// Unique returns a duplicate-free version of an array.
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

// Diff returns the difference between two slices.
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

// Pop returns an eent and removes it from the slice.
func Pop[T comparable](list *[]T) (v T) {
	l := len(*list)
	if l == 0 {
		return
	}

	v = (*list)[l-1]
	*list = (*list)[:l-1]
	return
}

// Shift returns an eent and removes it from the slice.
func Shift[T comparable](list *[]T) (v T) {
	l := len(*list)
	if l == 0 {
		return
	}

	v = (*list)[0]
	*list = (*list)[1:]
	return
}

// Chunk split slice into n parts
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

// RandShift returns a function that returns a random element from the list and removes it.
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
