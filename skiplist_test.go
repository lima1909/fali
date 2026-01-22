package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase(t *testing.T) {
	sl := NewSkipList[string]()
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

	assert.False(t, sl.Delete(2))

	val, found = sl.Get(1)
	assert.True(t, found)
	assert.Equal(t, "a", val)

	val, found = sl.Get(3)
	assert.True(t, found)
	assert.Equal(t, "c", val)
}

func TestRange(t *testing.T) {
	sl := NewSkipList[uint32]()
	sl.Put(1, 1)
	sl.Put(3, 3)
	sl.Put(5, 5)
	sl.Put(4, 4)

	result := sl.Range(2, 42)
	assert.Equal(t, []uint32{3, 4, 5}, result)
}

func TestRangeInclusiveTo(t *testing.T) {
	sl := NewSkipList[uint32]()
	sl.Put(1, 1)
	sl.Put(3, 3)
	sl.Put(5, 5)
	sl.Put(4, 4)

	result := sl.Range(2, 5)
	assert.Equal(t, []uint32{3, 4, 5}, result)
}

func TestRangeInclusiveFromTo(t *testing.T) {
	sl := NewSkipList[uint32]()
	sl.Put(2, 2)
	sl.Put(3, 3)
	sl.Put(5, 5)
	sl.Put(4, 4)

	result := sl.Range(2, 5)
	assert.Equal(t, []uint32{2, 3, 4, 5}, result)
}

func TestNotInRange(t *testing.T) {
	sl := NewSkipList[uint32]()
	sl.Put(1, 1)
	sl.Put(3, 3)

	result := sl.Range(4, 42)
	assert.Nil(t, result)
}

func BenchmarkSkiplist(b *testing.B) {
	count := 3_000_000
	found_val := 990_000

	sl := NewSkipList[uint32]()
	for i := 1; i <= count; i++ {
		sl.Put(uint32(i), uint32(i))
	}
	b.ResetTimer()

	for b.Loop() {
		_, found := sl.Get(uint32(found_val))
		if !found {
			panic(fmt.Sprintf("NOT FOUND: %d", found_val))
		}
	}
}
