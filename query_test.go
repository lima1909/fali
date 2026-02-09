package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func set[T any](idx Index32[T], t T, r uint32)   { idx.Set(&t, r) }
func unSet[T any](idx Index32[T], t T, r uint32) { idx.UnSet(&t, r) }

func stringGetFn(t *string) string { return *t }
func intGetFn(t *int) int          { return *t }

func fieldIndexMapFn[T any](mi Index32[T]) FieldIndexFn[uint32] {
	return func(fieldName string, _ any) (QueryFieldGetFn[uint32], error) {
		if fieldName == "val" {
			return mi.Get, nil
		}

		return nil, ErrInvalidIndexdName{fieldName}
	}
}

func TestMapIndex_UnSet(t *testing.T) {
	mi := NewMapIndex(intGetFn)
	set(mi, 1, 1)
	set(mi, 3, 3)
	set(mi, 3, 5)
	set(mi, 42, 42)

	// check all values are correct
	bs, err := mi.Get(Equal, 1)
	assert.NoError(t, err)
	assert.Equal(t, 1, bs.Count())
	bs, err = mi.Get(Equal, 3)
	assert.NoError(t, err)
	assert.Equal(t, 2, bs.Count())
	bs, err = mi.Get(Equal, 42)
	assert.NoError(t, err)
	assert.Equal(t, 1, bs.Count())

	// remove the last one: 42
	unSet(mi, 42, 42)
	_, err = mi.Get(Equal, 42)
	assert.ErrorIs(t, ErrValueNotFound{42}, err)

	// remove value 3
	unSet(mi, 3, 3)
	bs, err = mi.Get(Equal, 3)
	assert.NoError(t, err)
	assert.Equal(t, 1, bs.Count())
	unSet(mi, 3, 5)
	_, err = mi.Get(Equal, 3)
	assert.ErrorIs(t, ErrValueNotFound{3}, err)

	// for value 1 is no row 99, no deletion (ignored)
	unSet(mi, 1, 99)
	bs, err = mi.Get(Equal, 1)
	assert.NoError(t, err)
	assert.Equal(t, 1, bs.Count())

	// remove value 1
	unSet(mi, 1, 1)
	_, err = mi.Get(Equal, 1)
	assert.ErrorIs(t, ErrValueNotFound{1}, err)
}

func TestMapIndex_Get(t *testing.T) {
	mi := NewMapIndex(intGetFn)
	set(mi, 1, 1)
	set(mi, 3, 3)
	set(mi, 3, 5)
	set(mi, 42, 42)

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
	mi := NewMapIndex(intGetFn)
	set(mi, 1, 1)
	set(mi, 3, 3)
	set(mi, 3, 5)
	set(mi, 42, 42)

	fi := fieldIndexMapFn(mi)

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
	mi := NewMapIndex(intGetFn)
	set(mi, 1, 1)
	set(mi, 3, 3)
	set(mi, 3, 5)
	set(mi, 42, 42)

	fi := fieldIndexMapFn(mi)

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
	mi := NewMapIndex(intGetFn)
	set(mi, 1, 1)
	set(mi, 3, 3)
	set(mi, 3, 5)
	set(mi, 42, 42)

	fi := fieldIndexMapFn(mi)

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
	mi := NewMapIndex(intGetFn)
	set(mi, 1, 1)
	set(mi, 3, 3)
	set(mi, 3, 5)
	set(mi, 42, 42)

	fi := fieldIndexMapFn(mi)
	result, canMutate, err := All()(fi, NewBitSetFrom[uint32](1, 3, 5, 42))
	assert.NoError(t, err)
	assert.False(t, canMutate)
	assert.Equal(t, []uint32{1, 3, 5, 42}, result.ToSlice())
}
