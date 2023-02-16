//go:build go1.18
// +build go1.18

package zarray

type SortMaper[K hashable, V any] struct {
	values *Maper[K, V]
	keys   []K
}

func NewSortMap[K hashable, V any](size ...uintptr) *SortMaper[K, V] {
	return &SortMaper[K, V]{
		keys:   make([]K, 0),
		values: NewHashMap[K, V](size...),
	}
}

func (s *SortMaper[K, V]) Set(key K, value V) {
	if !s.values.Has(key) {
		s.keys = append(s.keys, key)
	}
	s.values.Set(key, value)
}

func (s *SortMaper[K, V]) Get(key K) (value V, ok bool) {
	return s.values.Get(key)
}

func (s *SortMaper[K, V]) Has(key K) (ok bool) {
	for _, v := range s.keys {
		if v == key {
			return true
		}
	}
	return false
}

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

func (s *SortMaper[K, V]) Len() int {
	return len(s.keys)
}

func (s *SortMaper[K, V]) Keys() []K {
	return s.keys
}

func (s *SortMaper[K, V]) ForEach(lambda func(K, V) bool) {
	for i := range s.keys {
		v, ok := s.values.Get(s.keys[i])
		if ok && lambda(s.keys[i], v) {
			continue
		}
		break
	}
}
