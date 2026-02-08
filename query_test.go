package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func fieldIndexMap(mi Index32[int]) FieldIndexFn[uint32] {
	return func(fieldName string, _ any) (QueryFieldGetFn[uint32], error) {
		if fieldName == "val" {
			return mi.Get, nil
		}

		return nil, ErrInvalidIndexdName{fieldName}
	}
}

func TestMapIndex_Set_UnSet(t *testing.T) {
	mi := NewMapIndex(func(t int) int { return t })
	mi.Set(1, 1)
	mi.Set(3, 3)
	mi.Set(3, 5)
	mi.Set(42, 42)
	assert.Equal(t, 3, len(mi.data))

	mi.UnSet(42, 42)
	assert.Equal(t, 2, len(mi.data))

	// the same len, because for key 3 still exist the row 5
	mi.UnSet(3, 3)
	assert.Equal(t, 2, len(mi.data))

	// for key 1 is no row 99, no deletion
	mi.UnSet(1, 99)
	assert.Equal(t, 2, len(mi.data))

	mi.UnSet(1, 1)
	assert.Equal(t, 1, len(mi.data))

	// map is empty
	mi.UnSet(3, 5)
	assert.Equal(t, 0, len(mi.data))
}

func TestMapIndex_Get(t *testing.T) {
	mi := NewMapIndex(func(t int) int { return t })
	mi.Set(1, 1)
	mi.Set(3, 3)
	mi.Set(3, 5)
	mi.Set(42, 42)

	bs, _ := mi.Get(Equal, 1)
	assert.Equal(t, NewBitSetFrom[uint32](1), bs)
	bs, _ = mi.Get(Equal, 3)
	assert.Equal(t, []uint32{3, 5}, bs.ToSlice())

	// not found
	_, err := mi.Get(Equal, 7)
	assert.ErrorIs(t, ErrValueNotFound{7}, err)
	// invalid relation
	_, err = mi.Get(Greater, 1)
	assert.ErrorIs(t, ErrInvalidRelation{Greater}, err)
}

func TestMapIndex_Query(t *testing.T) {
	mi := NewMapIndex(func(t int) int { return t })
	mi.Set(1, 1)
	mi.Set(3, 3)
	mi.Set(3, 5)
	mi.Set(42, 42)

	fi := fieldIndexMap(mi)

	result, canMutate, err := Eq("val", 3)(fi, nil)
	assert.NoError(t, err)
	assert.False(t, canMutate)
	assert.Equal(t, []uint32{3, 5}, result.ToSlice())

	// repeat the Eq with the same paramter, to check the result BitSet is not changed
	result, _, err = Eq("val", 3)(fi, nil)
	assert.NoError(t, err)
	assert.Equal(t, []uint32{3, 5}, result.ToSlice())

	// not found
	result, _, err = Eq("val", 99)(fi, nil)
	assert.ErrorIs(t, ErrValueNotFound{99}, err)
	assert.Nil(t, result)

	// invalid field
	result, _, err = Eq("bad", 1)(fi, nil)
	assert.ErrorIs(t, ErrInvalidIndexdName{"bad"}, err)
	assert.Nil(t, result)

	// OR
	result, canMutate, err =
		Eq("val", 3).
			Or(Eq("val", 1))(fi, nil)
	assert.NoError(t, err)
	assert.True(t, canMutate)
	assert.Equal(t, []uint32{1, 3, 5}, result.ToSlice())

	// And
	result, canMutate, err =
		Eq("val", 3).
			And(Eq("val", 3))(fi, nil)
	assert.NoError(t, err)
	assert.True(t, canMutate)
	assert.Equal(t, []uint32{3, 5}, result.ToSlice())

	// after and | or, to check the original BitSet is not changed
	bs, _ := mi.Get(Equal, 1)
	assert.Equal(t, []uint32{1}, bs.ToSlice())
	bs, _ = mi.Get(Equal, 42)
	assert.Equal(t, []uint32{42}, bs.ToSlice())
	bs, _ = mi.Get(Equal, 3)
	assert.Equal(t, []uint32{3, 5}, bs.ToSlice())
}

func TestMapIndex_Query_Not(t *testing.T) {
	mi := NewMapIndex(func(t int) int { return t })
	mi.Set(1, 1)
	mi.Set(3, 3)
	mi.Set(3, 5)
	mi.Set(42, 42)

	fi := fieldIndexMap(mi)

	allIDs := NewBitSetFrom[uint32](1, 3, 5, 42)

	// Not
	result, canMutate, err := Not(Eq("val", 3))(fi, allIDs)
	assert.NoError(t, err)
	assert.True(t, canMutate)
	assert.Equal(t, []uint32{1, 42}, result.ToSlice())

	// NotEq
	result, canMutate, err = NotEq("val", 3)(fi, allIDs)
	assert.NoError(t, err)
	assert.True(t, canMutate)
	assert.Equal(t, []uint32{1, 42}, result.ToSlice())

	// after and | or, to check the original BitSet is not changed
	bs, _ := mi.Get(Equal, 1)
	assert.Equal(t, []uint32{1}, bs.ToSlice())
	bs, _ = mi.Get(Equal, 42)
	assert.Equal(t, []uint32{42}, bs.ToSlice())
	bs, _ = mi.Get(Equal, 3)
	assert.Equal(t, []uint32{3, 5}, bs.ToSlice())
}

func TestMapIndex_Query_In(t *testing.T) {
	mi := NewMapIndex(func(t int) int { return t })
	mi.Set(1, 1)
	mi.Set(3, 3)
	mi.Set(3, 5)
	mi.Set(42, 42)

	fi := fieldIndexMap(mi)

	// In empty
	result, canMutate, err := In("val")(fi, nil)
	assert.NoError(t, err)
	assert.True(t, canMutate)
	assert.Equal(t, []uint32{}, result.ToSlice())

	// In one
	result, canMutate, err = In("val", 1)(fi, nil)
	assert.NoError(t, err)
	assert.False(t, canMutate)
	assert.Equal(t, []uint32{1}, result.ToSlice())

	// In many
	result, canMutate, err = In("val", 42, 1)(fi, nil)
	assert.NoError(t, err)
	assert.True(t, canMutate)
	assert.Equal(t, []uint32{1, 42}, result.ToSlice())

	// after and | or, to check the original BitSet is not changed
	bs, _ := mi.Get(Equal, 1)
	assert.Equal(t, []uint32{1}, bs.ToSlice())
	bs, _ = mi.Get(Equal, 42)
	assert.Equal(t, []uint32{42}, bs.ToSlice())
	bs, _ = mi.Get(Equal, 3)
	assert.Equal(t, []uint32{3, 5}, bs.ToSlice())
}

func TestMapIndex_QueryAll(t *testing.T) {
	mi := NewMapIndex(func(t int) int { return t })
	mi.Set(1, 1)
	mi.Set(3, 3)
	mi.Set(3, 5)
	mi.Set(42, 42)

	fi := fieldIndexMap(mi)
	result, canMutate, err := All()(fi, NewBitSetFrom[uint32](1, 3, 5, 42))
	assert.NoError(t, err)
	assert.False(t, canMutate)
	assert.Equal(t, []uint32{1, 3, 5, 42}, result.ToSlice())
}
