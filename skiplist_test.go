package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitListBase(t *testing.T) {
	sl := NewSkipList[int, string]()
	assert.True(t, sl.Put(1, "a"))
	assert.True(t, sl.Put(3, "c"))
	assert.True(t, sl.Put(2, "b"))
	assert.False(t, sl.Put(2, "b"))

	val, found := sl.Get(2)
	assert.True(t, found)
	assert.Equal(t, "b", val)

	assert.True(t, sl.Delete(2))
	val, found = sl.Get(2)
	assert.False(t, found)
	assert.Equal(t, "", val)

	assert.False(t, sl.Delete(2))

	val, found = sl.Get(1)
	assert.True(t, found)
	assert.Equal(t, "a", val)

	val, found = sl.Get(3)
	assert.True(t, found)
	assert.Equal(t, "c", val)
}

func TestNilValue(t *testing.T) {
	sl := NewSkipList[string, *string]()
	sl.Put("a", nil)

	val, found := sl.Get("a")
	assert.True(t, found)
	assert.Nil(t, val)
}

func TestPutWithZeroValueKey(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Expected panic when putting zero value as Key")
		}
	}()

	sl := NewSkipList[string, string]()
	sl.Put("", "---")
}

func TestDeleteAndGetTheZeroValueKey(t *testing.T) {
	sl := NewSkipList[string, string]()
	assert.False(t, sl.Delete(""))

	val, found := sl.Get("")
	assert.False(t, found)
	assert.Equal(t, "", val)
}

func TestTraverse(t *testing.T) {
	count := 10

	sl := NewSkipList[uint32, uint32]()
	for i := 1; i <= count; i++ {
		sl.Put(uint32(i), uint32(i))
	}

	c := 0
	toTheEnd := sl.Traverse(func(key, val uint32) bool {
		c += 1
		return true
	})
	assert.True(t, toTheEnd)
	assert.Equal(t, count, c)

	c = 0
	toTheEnd = sl.Traverse(func(key, val uint32) bool {
		c += 1
		return c != 5
	})
	assert.False(t, toTheEnd)
	assert.Equal(t, 5, c)

}

func TestRange(t *testing.T) {
	sl := NewSkipList[byte, uint32]()
	sl.Put(1, 1)
	sl.Put(3, 3)
	sl.Put(5, 5)
	sl.Put(4, 4)

	result := make([]uint32, 0)
	sl.Range(2, 42,
		func(key byte, val uint32) bool {
			result = append(result, val)
			return true
		})
	assert.Equal(t, []uint32{3, 4, 5}, result)
}

func TestRangeInclusiveTo(t *testing.T) {
	sl := NewSkipList[string, uint32]()
	sl.Put("a", 1)
	sl.Put("c", 3)
	sl.Put("z", 5)
	sl.Put("x", 4)

	result := make([]uint32, 0)
	sl.Range("b", "z",
		func(key string, val uint32) bool {
			result = append(result, val)
			return true
		})
	assert.Equal(t, []uint32{3, 4, 5}, result)
}

func TestRangeInclusiveFromTo(t *testing.T) {
	sl := NewSkipList[int, uint32]()
	sl.Put(2, 2)
	sl.Put(3, 3)
	sl.Put(5, 5)
	sl.Put(4, 4)

	result := make([]uint32, 0)
	sl.Range(2, 5,
		func(key int, val uint32) bool {
			result = append(result, val)
			return true
		})
	assert.Equal(t, []uint32{2, 3, 4, 5}, result)
}

func TestNotInRange(t *testing.T) {
	sl := NewSkipList[uint32, uint32]()
	sl.Put(1, 1)
	sl.Put(3, 3)

	result := make([]uint32, 0)
	sl.Range(4, 42,
		func(key, val uint32) bool {
			result = append(result, val)
			return true
		})
	assert.Equal(t, 0, len(result))
}

func TestFirstValue(t *testing.T) {
	sl := NewSkipList[uint32, uint32]()
	val, ok := sl.FirstValue()
	assert.False(t, ok)
	assert.Equal(t, uint32(0), val)

	sl.Put(1, 1)
	val, ok = sl.FirstValue()
	assert.True(t, ok)
	assert.Equal(t, uint32(1), val)
}

func TestLastValue(t *testing.T) {
	sl := NewSkipList[uint32, uint32]()
	val, ok := sl.LastValue()
	assert.False(t, ok)
	assert.Equal(t, uint32(0), val)

	sl.Put(1, 1)
	val, ok = sl.LastValue()
	assert.True(t, ok)
	assert.Equal(t, uint32(1), val)

	sl.Put(5, 5)
	val, ok = sl.LastValue()
	assert.True(t, ok)
	assert.Equal(t, uint32(5), val)
}

func TestMinKey(t *testing.T) {
	sl := NewSkipList[int, uint32]()
	sl.Put(1, 2)
	sl.Put(3, 4)

	k, found := sl.MinKey()
	assert.True(t, found)
	assert.Equal(t, 1, k)

	sl.Delete(1)
	k, found = sl.MinKey()
	assert.True(t, found)
	assert.Equal(t, 3, k)

	sl.Delete(3)
	k, found = sl.MinKey()
	assert.False(t, found)
	assert.Equal(t, 0, k)
}

func TestMaxKey(t *testing.T) {
	sl := NewSkipList[int, uint32]()
	sl.Put(1, 2)
	sl.Put(3, 4)

	k, found := sl.MaxKey()
	assert.True(t, found)
	assert.Equal(t, 3, k)

	sl.Delete(3)
	k, found = sl.MaxKey()
	assert.True(t, found)
	assert.Equal(t, 1, k)

	sl.Delete(1)
	k, found = sl.MaxKey()
	assert.False(t, found)
	assert.Equal(t, 0, k)
}
