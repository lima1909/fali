package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortedIndex_Equal(t *testing.T) {
	si := NewSortedIndex[uint16]()
	si.Set(1, 1)
	si.Set(1, 2)
	si.Set(3, 3)

	assert.Equal(t, []uint16{1, 2}, si.Get(Equal, 1).ToSlice())

	si.UnSet(1, 2)
	assert.Equal(t, []uint16{1}, si.Get(Equal, 1).ToSlice())
	si.UnSet(1, 1)
	assert.Equal(t, []uint16{}, si.Get(Equal, 1).ToSlice())
}

func TestSortedIndex_Less(t *testing.T) {
	si := NewSortedIndex[uint16]()
	si.Set(1, 1)
	si.Set(1, 2)
	si.Set(3, 3)

	assert.Equal(t, []uint16{}, si.Get(Less, 0).ToSlice())
	assert.Equal(t, []uint16{}, si.Get(Less, 1).ToSlice())
	assert.Equal(t, []uint16{1, 2}, si.Get(Less, 2).ToSlice())
	assert.Equal(t, []uint16{1, 2}, si.Get(Less, 3).ToSlice())
	assert.Equal(t, []uint16{1, 2, 3}, si.Get(Less, 5).ToSlice())

}

func TestSortedIndex_LessEqual(t *testing.T) {
	si := NewSortedIndex[uint16]()
	si.Set(1, 1)
	si.Set(1, 2)
	si.Set(3, 3)

	assert.Equal(t, []uint16{}, si.Get(LessEqual, 0).ToSlice())
	assert.Equal(t, []uint16{1, 2}, si.Get(LessEqual, 1).ToSlice())
	assert.Equal(t, []uint16{1, 2}, si.Get(LessEqual, 2).ToSlice())
	assert.Equal(t, []uint16{1, 2, 3}, si.Get(LessEqual, 3).ToSlice())
	assert.Equal(t, []uint16{1, 2, 3}, si.Get(LessEqual, 5).ToSlice())

}
