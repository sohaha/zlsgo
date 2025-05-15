//go:build go1.18
// +build go1.18

package zarray

// SortMaper implements an ordered map that maintains insertion order of keys
// while providing map-like operations for key-value pairs.
type SortMaper[K hashable, V any] struct {
	values *Maper[K, V]
	keys   []K
}

// NewSortMap creates a new SortMaper with the specified initial capacity.
// The SortMaper maintains the insertion order of keys while providing map operations.
func NewSortMap[K hashable, V any](size ...uintptr) *SortMaper[K, V] {
	return &SortMaper[K, V]{
		keys:   make([]K, 0),
		values: NewHashMap[K, V](size...),
	}
}

// Set adds or updates a key-value pair in the map.
// If the key is new, it is appended to the ordered keys list.
func (s *SortMaper[K, V]) Set(key K, value V) {
	if !s.values.Has(key) {
		s.keys = append(s.keys, key)
	}
	s.values.Set(key, value)
}

// Get retrieves a value by its key.
// Returns the value and a boolean indicating whether the key was found.
func (s *SortMaper[K, V]) Get(key K) (value V, ok bool) {
	return s.values.Get(key)
}

// Has checks if a key exists in the map.
// Returns true if the key exists, false otherwise.
func (s *SortMaper[K, V]) Has(key K) (ok bool) {
	for _, v := range s.keys {
		if v == key {
			return true
		}
	}
	return false
}

// Delete removes one or more key-value pairs from the map.
// This removes the keys from both the ordered keys list and the underlying map.
func (s *SortMaper[K, V]) Delete(key ...K) {
	for i, v := range s.keys {
		for _, k := range key {
			if v == k {
				s.keys = append(s.keys[:i], s.keys[i+1:]...)
				break
			}
		}
	}
	s.values.Delete(key...)
}

// Len returns the number of key-value pairs in the map.
func (s *SortMaper[K, V]) Len() int {
	return len(s.keys)
}

// Keys returns all keys in the map in their insertion order.
func (s *SortMaper[K, V]) Keys() []K {
	return s.keys
}

// ForEach iterates through all key-value pairs in the map in insertion order.
// The iteration continues as long as the lambda function returns true.
func (s *SortMaper[K, V]) ForEach(lambda func(K, V) bool) {
	for i := range s.keys {
		v, ok := s.values.Get(s.keys[i])
		if ok && lambda(s.keys[i], v) {
			continue
		}
		break
	}
}
