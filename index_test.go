package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortedIndex_Equal(t *testing.T) {
	si := NewSortedIndex(SelfFn[string]())
	set(si, "a", 1)
	set(si, "a", 2)
	set(si, "b", 3)

	bs, _ := si.Get(Equal, "a")
	assert.Equal(t, []uint32{1, 2}, bs.ToSlice())

	unSet(si, "a", 2)
	bs, _ = si.Get(Equal, "a")
	assert.Equal(t, []uint32{1}, bs.ToSlice())

	unSet(si, "a", 1)
	_, err := si.Get(Equal, "a")
	assert.ErrorIs(t, ErrValueNotFound{"a"}, err)
}

func TestSortedIndex_Less(t *testing.T) {
	si := NewSortedIndex(SelfFn[int]())
	set(si, 1, 1)
	set(si, 1, 2)
	set(si, 3, 3)

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
	si := NewSortedIndex(SelfFn[int]())
	set(si, 1, 1)
	set(si, 1, 2)
	set(si, 3, 3)

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

func TestIDIndex_Lookup(t *testing.T) {
	mi := newIDMapIndex((*car).Name)
	vw := car{name: "vw", age: 2}
	mi.Set(&vw, 0)

	bs, err := mi.Get(Equal, "vw")
	assert.NoError(t, err)
	assert.Equal(t, []uint32{0}, bs.ToSlice())

	_, err = mi.Get(Equal, 4)
	assert.ErrorIs(t, ErrInvalidIndexValue[string]{4}, err)

	_, err = mi.Get(Less, "vw")
	assert.ErrorIs(t, ErrInvalidRelation{Less}, err)

	_, err = mi.Get(Equal, "opel")
	assert.ErrorIs(t, ErrValueNotFound{"opel"}, err)
}
