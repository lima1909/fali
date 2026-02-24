package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortedIndex_Equal(t *testing.T) {
	si := NewSortedIndex(FromValue[string]())
	set(si, "a", 1)
	set(si, "a", 2)
	set(si, "b", 3)

	bs, _ := si.Get(OpEq, "a")
	assert.Equal(t, []uint32{1, 2}, bs.ToSlice())

	unSet(si, "a", 2)
	bs, _ = si.Get(OpEq, "a")
	assert.Equal(t, []uint32{1}, bs.ToSlice())

	unSet(si, "a", 1)
	bs, err := si.Get(OpEq, "a")
	assert.NoError(t, err)
	assert.Equal(t, 0, bs.Count())
}

func TestSortedIndex_Less(t *testing.T) {
	si := NewSortedIndex(FromValue[int]())
	set(si, 1, 1)
	set(si, 1, 2)
	set(si, 3, 3)

	bs, _ := si.Get(OpLt, 0)
	assert.Equal(t, []uint32{}, bs.ToSlice())
	bs, _ = si.Get(OpLt, 1)
	assert.Equal(t, []uint32{}, bs.ToSlice())
	bs, _ = si.Get(OpLt, 2)
	assert.Equal(t, []uint32{1, 2}, bs.ToSlice())
	bs, _ = si.Get(OpLt, 3)
	assert.Equal(t, []uint32{1, 2}, bs.ToSlice())
	bs, _ = si.Get(OpLt, 5)
	assert.Equal(t, []uint32{1, 2, 3}, bs.ToSlice())
}

func TestSortedIndex_LessEqual(t *testing.T) {
	si := NewSortedIndex(FromValue[int]())
	set(si, 1, 1)
	set(si, 1, 2)
	set(si, 3, 3)

	bs, _ := si.Get(OpLe, 0)
	assert.Equal(t, []uint32{}, bs.ToSlice())
	bs, _ = si.Get(OpLe, 1)
	assert.Equal(t, []uint32{1, 2}, bs.ToSlice())
	bs, _ = si.Get(OpLe, 2)
	assert.Equal(t, []uint32{1, 2}, bs.ToSlice())
	bs, _ = si.Get(OpLe, 3)
	assert.Equal(t, []uint32{1, 2, 3}, bs.ToSlice())
	bs, _ = si.Get(OpLe, 5)
	assert.Equal(t, []uint32{1, 2, 3}, bs.ToSlice())
}

func TestIDIndex_Lookup(t *testing.T) {
	mi := newIDMapIndex((*car).Name)
	vw := car{name: "vw", age: 2}
	mi.Set(&vw, 0)

	bs, err := mi.Get(OpEq, "vw")
	assert.NoError(t, err)
	assert.Equal(t, []uint32{0}, bs.ToSlice())

	_, err = mi.Get(OpEq, 4)
	assert.ErrorIs(t, ErrInvalidIndexValue[string]{4}, err)

	_, err = mi.Get(OpLt, "vw")
	assert.ErrorIs(t, ErrInvalidOperation{OpLt}, err)

	_, err = mi.Get(OpEq, "opel")
	assert.ErrorIs(t, ErrValueNotFound{"opel"}, err)
}
