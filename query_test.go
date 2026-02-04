package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func fieldIndexMap(mi Index[uint16]) FieldIndexFn[uint16] {
	return func(fieldName string, _ any) (Index[uint16], error) {
		if fieldName == "val" {
			return mi, nil
		}
		return nil, fmt.Errorf("not found: %s", fieldName)
	}
}

func TestMapIndex_Set_UnSet(t *testing.T) {
	mi := NewMapIndex[uint16]()
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
	mi := NewMapIndex[uint16]()
	mi.Set(1, 1)
	mi.Set(3, 3)
	mi.Set(3, 5)
	mi.Set(42, 42)

	assert.Equal(t, NewBitSetFrom[uint16](1), mi.Get(Equal, 1))
	assert.Equal(t, []uint16{3, 5}, mi.Get(Equal, 3).ToSlice())

	// not found
	assert.Equal(t, NewBitSet[uint16](), mi.Get(Equal, 99))
	// invalid relation
	assert.Equal(t, NewBitSet[uint16](), mi.Get(Greater, 1))
}

func TestMapIndex_Query(t *testing.T) {
	mi := NewMapIndex[uint16]()
	mi.Set(1, 1)
	mi.Set(3, 3)
	mi.Set(3, 5)
	mi.Set(42, 42)

	fi := fieldIndexMap(mi)

	result, canMutate, err := Eq[uint16]("val", 3)(fi, nil)
	assert.NoError(t, err)
	assert.False(t, canMutate)
	assert.Equal(t, []uint16{3, 5}, result.ToSlice())

	// repeat the Eq with the same paramter, to check the result BitSet is not changed
	result, _, err = Eq[uint16]("val", 3)(fi, nil)
	assert.NoError(t, err)
	assert.Equal(t, []uint16{3, 5}, result.ToSlice())

	// not found
	result, _, err = Eq[uint16]("val", 99)(fi, nil)
	// test function doesn't throw an error!
	assert.NoError(t, err)
	assert.Equal(t, []uint16{}, result.ToSlice())
	// invalid field
	result, _, err = Eq[uint16]("bad", 99)(fi, nil)
	assert.Error(t, err)
	assert.Nil(t, result)

	// OR
	result, canMutate, err =
		Eq[uint16]("val", 3).
			Or(Eq[uint16]("val", 1))(fi, nil)
	assert.NoError(t, err)
	assert.True(t, canMutate)
	assert.Equal(t, []uint16{1, 3, 5}, result.ToSlice())

	// And
	result, canMutate, err =
		Eq[uint16]("val", 3).
			And(Eq[uint16]("val", 3))(fi, nil)
	assert.NoError(t, err)
	assert.True(t, canMutate)
	assert.Equal(t, []uint16{3, 5}, result.ToSlice())

	// after and | or, to check the original BitSet is not changed
	assert.Equal(t, []uint16{1}, mi.Get(Equal, 1).ToSlice())
	assert.Equal(t, []uint16{42}, mi.Get(Equal, 42).ToSlice())
	assert.Equal(t, []uint16{3, 5}, mi.Get(Equal, 3).ToSlice())
}

func TestMapIndex_Query_Not(t *testing.T) {
	mi := NewMapIndex[uint16]()
	mi.Set(1, 1)
	mi.Set(3, 3)
	mi.Set(3, 5)
	mi.Set(42, 42)

	fi := fieldIndexMap(mi)

	allIDs := NewBitSetFrom[uint16](1, 3, 5, 42)

	// Not
	result, canMutate, err := Not(Eq[uint16]("val", 3))(fi, allIDs)
	assert.NoError(t, err)
	assert.True(t, canMutate)
	assert.Equal(t, []uint16{1, 42}, result.ToSlice())

	// NotEq
	result, canMutate, err = NotEq[uint16]("val", 3)(fi, allIDs)
	assert.NoError(t, err)
	assert.True(t, canMutate)
	assert.Equal(t, []uint16{1, 42}, result.ToSlice())

	// after and | or, to check the original BitSet is not changed
	assert.Equal(t, []uint16{1}, mi.Get(Equal, 1).ToSlice())
	assert.Equal(t, []uint16{42}, mi.Get(Equal, 42).ToSlice())
	assert.Equal(t, []uint16{3, 5}, mi.Get(Equal, 3).ToSlice())
}

func TestMapIndex_Query_In(t *testing.T) {
	mi := NewMapIndex[uint16]()
	mi.Set(1, 1)
	mi.Set(3, 3)
	mi.Set(3, 5)
	mi.Set(42, 42)

	fi := fieldIndexMap(mi)

	// In empty
	result, canMutate, err := In[uint16]("val")(fi, nil)
	assert.NoError(t, err)
	assert.True(t, canMutate)
	assert.Equal(t, []uint16{}, result.ToSlice())

	// In one
	result, canMutate, err = In[uint16]("val", 1)(fi, nil)
	assert.NoError(t, err)
	assert.False(t, canMutate)
	assert.Equal(t, []uint16{1}, result.ToSlice())

	// In many
	result, canMutate, err = In[uint16]("val", 42, 1)(fi, nil)
	assert.NoError(t, err)
	assert.True(t, canMutate)
	assert.Equal(t, []uint16{1, 42}, result.ToSlice())

	// after and | or, to check the original BitSet is not changed
	assert.Equal(t, []uint16{1}, mi.Get(Equal, 1).ToSlice())
	assert.Equal(t, []uint16{42}, mi.Get(Equal, 42).ToSlice())
	assert.Equal(t, []uint16{3, 5}, mi.Get(Equal, 3).ToSlice())
}
