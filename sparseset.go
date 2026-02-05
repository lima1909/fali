package main

import (
	"slices"
)

type SparseSet[V Value] struct {
	data []V
}

func NewSparseSet[V Value]() *SparseSet[V] {
	return &SparseSet[V]{data: make([]V, 0)}
}

func (s *SparseSet[V]) Set(value V) {
	if len(s.data) == 0 || value > s.data[len(s.data)-1] {
		s.data = append(s.data, value)
		return
	}

	s.data = append(s.data, value)
	slices.Sort(s.data)
	s.data = slices.Compact(s.data)
}

func (s *SparseSet[V]) UnSet(value V) bool {
	idx, found := slices.BinarySearch(s.data, value)
	if !found {
		return false
	}

	s.data = append(s.data[:idx], s.data[idx+1:]...)
	return true

}

func (s *SparseSet[V]) Contains(value V) bool {
	_, found := slices.BinarySearch(s.data, value)
	return found
}

// Min return the min value of this set
// [1, 3, 100] => 1
// if the set is empty, return -1
func (s *SparseSet[V]) Min() int {
	if len(s.data) == 0 {
		return -1
	}

	return int(s.data[0])
}

// Max return the max value of this set
// [1, 3, 100] => 100
// if the set is empty, return -1
func (s *SparseSet[V]) Max() int {
	l := len(s.data)
	if l == 0 {
		return -1
	}

	return int(s.data[l-1])
}

func (s *SparseSet[V]) Count() int { return len(s.data) }
func (s *SparseSet[V]) Len() int   { return len(s.data) }

func (s *SparseSet[V]) Copy() *SparseSet[V] {
	target := make([]V, len(s.data))
	copy(target, s.data)
	return &SparseSet[V]{data: target}
}

func (s *SparseSet[V]) And2(other *SparseSet[V]) {
	sa := s.data
	so := other.data
	l := min(len(sa), len(so))

	for i := range l {
		sa[i] &= so[i]
	}

	s.data = sa[:l]
}

// func (s *SparseSet[V]) And(other *SparseSet[V]) {
// 	// if other == nil || len(other.data) == 0 {
// 	// 	s.data = s.data[:0]
// 	// 	return
// 	// }
// 	//
// 	// if s == other {
// 	// 	return
// 	// }
//
// 	if other == nil || s == other || (len(s.data) > 0 && len(other.data) > 0 && &s.data[0] == &other.data[0]) {
// 		return
// 	}
//
// 	sa := s.data
// 	so := other.data
// 	// if len(sa) > 0 && &sa[0] == &so[0] {
// 	// 	return
// 	// }
//
// 	l := min(len(sa), len(so))
// 	for i := range l {
// 		sa[i] &= so[i]
// 	}
//
// 	s.data = sa[:l]
// }

func (s *SparseSet[V]) Or(other *SparseSet[V]) {
	so := other.data
	if len(so) > len(s.data) {
		s.expand(len(so))
	}
	sa := s.data

	for i, val := range other.data {
		sa[i] |= val
	}
}

func (s *SparseSet[V]) Xor(other *SparseSet[V]) {
	so := other.data
	if len(so) > len(s.data) {
		s.expand(len(so))
	}
	sa := s.data

	for i, val := range so {
		sa[i] ^= val
	}
}

func (s *SparseSet[V]) AndNot(other *SparseSet[V]) {
	sa := s.data
	so := other.data
	l := min(len(sa), len(so))

	for i := range l {
		sa[i] &^= so[i]
	}
}

func (s *SparseSet[V]) Values(yield func(V) bool) {
	sa := s.data
	for _, v := range sa {
		if !yield(v) {
			return
		}
	}
}

func (s *SparseSet[V]) ToSlice() []V {
	return s.data
}

func (s *SparseSet[V]) ToBitSet() *BitSet[V] {
	return NewBitSetFrom(s.data...)
}

// expand ensures the BitSet has enough capacity to hold 'newLen' uint64s.
// It is inlined for speed.
func (s *SparseSet[V]) expand(newLen int) {
	if cap(s.data) >= newLen {
		// Fast path: Just change the length (data remains valid)
		s.data = s.data[:newLen]
	} else {
		// Slow path: Allocate larger array and copy
		newData := make([]V, newLen)
		copy(newData, s.data)
		s.data = newData
	}
}
