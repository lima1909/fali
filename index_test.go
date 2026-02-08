package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortedIndex_Equal(t *testing.T) {
	si := NewSortedIndex(func(t string) string { return t })
	si.Set("a", 1)
	si.Set("a", 2)
	si.Set("b", 3)

	bs, _ := si.Get(Equal, "a")
	assert.Equal(t, []uint32{1, 2}, bs.ToSlice())

	si.UnSet("a", 2)
	bs, _ = si.Get(Equal, "a")
	assert.Equal(t, []uint32{1}, bs.ToSlice())

	si.UnSet("a", 1)
	_, err := si.Get(Equal, "a")
	assert.ErrorIs(t, ErrValueNotFound{"a"}, err)
}

func TestSortedIndex_Less(t *testing.T) {
	si := NewSortedIndex(func(t int) int { return t })
	si.Set(1, 1)
	si.Set(1, 2)
	si.Set(3, 3)

	bs, _ := si.Get(Less, 0)
	assert.Equal(t, []uint32{}, bs.ToSlice())
	bs, _ = si.Get(Less, 1)
	assert.Equal(t, []uint32{}, bs.ToSlice())
	bs, _ = si.Get(Less, 2)
	assert.Equal(t, []uint32{1, 2}, bs.ToSlice())
	bs, _ = si.Get(Less, 3)
	assert.Equal(t, []uint32{1, 2}, bs.ToSlice())
	bs, _ = si.Get(Less, 5)
	assert.Equal(t, []uint32{1, 2, 3}, bs.ToSlice())
}

func TestSortedIndex_LessEqual(t *testing.T) {
	si := NewSortedIndex(func(t int) int { return t })
	si.Set(1, 1)
	si.Set(1, 2)
	si.Set(3, 3)

	bs, _ := si.Get(LessEqual, 0)
	assert.Equal(t, []uint32{}, bs.ToSlice())
	bs, _ = si.Get(LessEqual, 1)
	assert.Equal(t, []uint32{1, 2}, bs.ToSlice())
	bs, _ = si.Get(LessEqual, 2)
	assert.Equal(t, []uint32{1, 2}, bs.ToSlice())
	bs, _ = si.Get(LessEqual, 3)
	assert.Equal(t, []uint32{1, 2, 3}, bs.ToSlice())
	bs, _ = si.Get(LessEqual, 5)
	assert.Equal(t, []uint32{1, 2, 3}, bs.ToSlice())
}
