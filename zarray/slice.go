//go:build go1.18
// +build go1.18

package zarray

import (
	"math/rand"
)

// Shuffle creates a slice of shuffled values
func Shuffle[T any](collection []T) []T {
	l := len(collection)
	n := make([]T, l)
	copy(n, collection)
	rand.Shuffle(l, func(i, j int) {
		n[i], n[j] = n[j], n[i]
	})

	return n
}

// Filter iterates over elements of collection
func Filter[T any](slice []T, predicate func(index int, item T) bool) []T {
	l := len(slice)
	res := make([]T, 0, l)
	for i := 0; i < l; i++ {
		v := slice[i]
		if predicate(i, v) {
			res = append(res, v)
		}
	}
	return res
}

// Contains returns true if an element is present in a collection
func Contains[T comparable](collection []T, element T) bool {
	for _, item := range collection {
		if item == element {
			return true
		}
	}

	return false
}
