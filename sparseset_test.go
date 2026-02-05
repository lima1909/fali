package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSparseSet_Base(t *testing.T) {
	sp := NewSparseSet[uint16]()
	sp.Set(42)
	sp.Set(3)
	sp.Set(1)

	assert.Equal(t, 3, sp.Count())
	assert.Equal(t, 3, sp.Len())
	assert.Equal(t, 1, sp.Min())
	assert.Equal(t, 42, sp.Max())
	assert.Equal(t, []uint16{1, 3, 42}, sp.ToSlice())

}

// func TestSparseSet_And(t *testing.T) {
// 	s1 := NewSparseSet[uint16]()
// 	s1.Set(42)
// 	s1.Set(3)
// 	s1.Set(1)
//
// 	s2 := NewSparseSet[uint16]()
// 	s2.Set(2)
// 	s2.Set(3)
// 	s2.Set(0)
//
// 	result := s1.Copy()
// 	result.And(s2)
// 	assert.Equal(t, []uint16{3}, result.ToSlice())
//
// }
