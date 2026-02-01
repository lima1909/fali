package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapIndex_Set_UnSet(t *testing.T) {
	mi := NewMapIndex[uint16]()
	mi.Set(Int(1), 1)
	mi.Set(Int(3), 3)
	mi.Set(Int(3), 5)
	mi.Set(Int(42), 42)
	assert.Equal(t, 3, len(mi.data))

	mi.UnSet(Int(42), 42)
	assert.Equal(t, 2, len(mi.data))

	// the same len, because for key 3 still exist the row 5
	mi.UnSet(Int(3), 3)
	assert.Equal(t, 2, len(mi.data))

	// for key 1 is no row 99, no deletion
	mi.UnSet(Int(1), 99)
	assert.Equal(t, 2, len(mi.data))

	mi.UnSet(Int(1), 1)
	assert.Equal(t, 1, len(mi.data))

	// map is empty
	mi.UnSet(Int(3), 5)
	assert.Equal(t, 0, len(mi.data))
}

func TestMapIndex_Get(t *testing.T) {
	mi := NewMapIndex[uint16]()
	mi.Set(Int(1), 1)
	mi.Set(Int(3), 3)
	mi.Set(Int(3), 5)
	mi.Set(Int(42), 42)

	assert.Equal(t, NewBitSetFrom[uint16](1), mi.Get(Equal, Int(1)))
	assert.Equal(t, []uint16{3, 5}, mi.Get(Equal, Int(3)).ToSlice())

	// not found
	assert.Equal(t, NewBitSet[uint16](), mi.Get(Equal, Int(99)))
	// invalid relation
	assert.Equal(t, NewBitSet[uint16](), mi.Get(Greater, Int(1)))
}

func TestMapIndex_Query(t *testing.T) {
	mi := NewMapIndex[uint16]()
	mi.Set(Int(1), 1)
	mi.Set(Int(3), 3)
	mi.Set(Int(3), 5)
	mi.Set(Int(42), 42)

	var fi FieldIndex[uint16] = map[string]Index[uint16]{
		"val": mi,
	}

	result, canMutate := Eq[uint16]("val", Int(3))(fi, nil)
	assert.False(t, canMutate)
	assert.Equal(t, []uint16{3, 5}, result.ToSlice())

	// repeat the Eq with the same paramter, to check the result BitSet is not changed
	result, _ = Eq[uint16]("val", Int(3))(fi, nil)
	assert.Equal(t, []uint16{3, 5}, result.ToSlice())

	// not found
	result, _ = Eq[uint16]("val", Int(99))(fi, nil)
	assert.Equal(t, []uint16{}, result.ToSlice())
	// invalid field
	result, _ = Eq[uint16]("bad", Int(99))(fi, nil)
	assert.Equal(t, []uint16{}, result.ToSlice())

	// Not
	allIDs := NewBitSetFrom[uint16](1, 3, 5, 42)
	result, canMutate = Not(Eq[uint16]("val", Int(3)))(fi, allIDs)
	assert.True(t, canMutate)
	assert.Equal(t, []uint16{1, 42}, result.ToSlice())

	// OR
	result, canMutate =
		Eq[uint16]("val", Int(3)).
			Or(Eq[uint16]("val", Int(1)))(fi, nil)
	assert.True(t, canMutate)
	assert.Equal(t, []uint16{1, 3, 5}, result.ToSlice())

	// And
	result, canMutate =
		Eq[uint16]("val", Int(3)).
			And(Eq[uint16]("val", Int(3)))(fi, nil)
	assert.True(t, canMutate)
	assert.Equal(t, []uint16{3, 5}, result.ToSlice())

	// after and | or, to check the original BitSet is not changed
	assert.Equal(t, []uint16{1}, mi.Get(Equal, Int(1)).ToSlice())
	assert.Equal(t, []uint16{42}, mi.Get(Equal, Int(42)).ToSlice())
	assert.Equal(t, []uint16{3, 5}, mi.Get(Equal, Int(3)).ToSlice())
}
