package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBitSetBase(t *testing.T) {
	b := NewBitSet[uint8]()
	assert.False(t, b.Contains(0))

	b.Set(0)
	b.Set(1)
	b.Set(2)
	b.Set(2)
	b.Set(42)

	assert.True(t, b.Contains(0))

	assert.Equal(t, 4, b.Count())
	assert.Equal(t, 1, b.Len())
	assert.Equal(t, 42, b.MaxSetValue())
	assert.Equal(t, 0, b.MaxSetIndex())
	assert.True(t, b.Contains(2))

	b.UnSet(2)
	assert.False(t, b.Contains(2))
	assert.Equal(t, 3, b.Count())
	assert.Equal(t, 1, b.Len())

	b.UnSet(42)
	assert.Equal(t, 1, b.MaxSetValue())
}

func TestBitSetToBig(t *testing.T) {
	b := NewBitSet[uint32]()

	assert.Equal(t, -1, b.MaxSetIndex())

	assert.False(t, b.UnSet(40_000))
	assert.False(t, b.Contains(40_000))
}

func TestBitSetShrink(t *testing.T) {
	b := NewBitSet[uint16]()
	b.Set(1)
	b.Set(130)

	assert.Equal(t, 2, b.MaxSetIndex())

	assert.Equal(t, 2, b.Count())
	assert.Equal(t, 3, b.Len())
	assert.True(t, b.UnSet(130))

	b.Shrink()
	assert.Equal(t, 1, b.Count())
	assert.Equal(t, 1, b.Len())
	assert.Equal(t, 0, b.MaxSetIndex())
}

func TestBitSetAnd(t *testing.T) {
	b1 := NewBitSetFrom[uint32](1, 2, 110, 2345)
	b2 := NewBitSetFrom[uint32](110)
	result := b1.Copy()
	result.And(b2)
	assert.Equal(t, NewBitSetFrom[uint32](110), result)

	b1 = NewBitSetFrom[uint32](110)
	b2 = NewBitSetFrom[uint32](1, 2, 110, 2345)
	result = b1.Copy()
	result.And(b2)
	assert.Equal(t, NewBitSetFrom[uint32](110), result)
}

func TestBitSetOr(t *testing.T) {
	b1 := NewBitSetFrom[uint32](1, 2, 110, 2345)
	b2 := NewBitSetFrom[uint32](110)
	result := b1.Copy()
	result.Or(b2)
	assert.Equal(t, NewBitSetFrom[uint32](1, 2, 110, 2345), result)

	b1 = NewBitSetFrom[uint32](110)
	b2 = NewBitSetFrom[uint32](1, 2, 110, 2345)
	result = b1.Copy()
	result.Or(b2)
	assert.Equal(t, NewBitSetFrom[uint32](1, 2, 110, 2345), result)
}

func TestBitSetXor(t *testing.T) {
	b1 := NewBitSetFrom[uint32](1, 2, 110, 2345)
	b2 := NewBitSetFrom[uint32](110)
	result := b1.Copy()
	result.Xor(b2)
	assert.Equal(t, NewBitSetFrom[uint32](1, 2, 2345), result)

	// shrinked?
	// assert.Equal(t, 1, result.Count())
	// assert.Equal(t, 2, len(result.data))

	b1 = NewBitSetFrom[uint32](110)
	b2 = NewBitSetFrom[uint32](1, 2, 110, 2345)
	result = b1.Copy()
	result.Xor(b2)
	assert.Equal(t, NewBitSetFrom[uint32](1, 2, 2345), result)
}

func TestBitSetAndNot(t *testing.T) {
	b1 := NewBitSetFrom[uint64](1, 2, 110, 2345)
	b2 := NewBitSetFrom[uint64](110, 2)
	result := b1.Copy()
	result.AndNot(b2)
	assert.Equal(t, NewBitSetFrom[uint64](1, 2345), result)

	b1 = NewBitSetFrom[uint64](110, 2)
	b2 = NewBitSetFrom[uint64](1, 2, 110, 2345)
	result = b1.Copy()
	result.AndNot(b2)
	result.Shrink()
	assert.Equal(t, NewBitSetFrom[uint64](), result)
}
