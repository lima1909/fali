package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFreeListBase(t *testing.T) {
	l := NewFreeList[string]()
	assert.Equal(t, 0, l.Add("a"))
	assert.Equal(t, 1, l.Add("b"))
	assert.Equal(t, 2, l.Add("c"))

	val, found := l.Get(1)
	assert.True(t, found)
	assert.Equal(t, "b", val)

	assert.False(t, l.Remove(100))
	assert.True(t, l.Remove(1))

	val, found = l.Get(1)
	assert.False(t, found)
	assert.Equal(t, "", val)

	l.Add("z")
	val, found = l.Get(1)
	assert.True(t, found)
	assert.Equal(t, "z", val)
}

func TestFreeListCompactUnstable(t *testing.T) {
	l := NewFreeList[string]()
	l.Add("a")
	l.Add("b")
	l.Add("c")
	l.Add("d")
	l.Add("e")
	l.Add("f")

	l.Remove(1) // b
	l.Remove(2) // c
	l.Remove(4) // e

	l.CompactUnstable()
	assert.Equal(t, 3, len(l.slots))

	val, found := l.Get(0)
	assert.True(t, found)
	assert.Equal(t, "a", val)

	val, found = l.Get(1)
	assert.True(t, found)
	assert.Equal(t, "d", val)

	val, found = l.Get(2)
	assert.True(t, found)
	assert.Equal(t, "f", val)
}

func TestFreeListCompactLinear(t *testing.T) {
	l := NewFreeList[string]()
	l.Add("a")
	l.Add("b")
	l.Add("c")
	l.Add("d")
	l.Add("e")
	l.Add("f")

	l.Remove(1) // b
	l.Remove(2) // c
	l.Remove(4) // e

	removed := make([]int, 0)
	l.CompactLinear(func(oldIndex, newIndex int) {
		removed = append(removed, oldIndex)
	})
	// the index 0 is not moved
	assert.Equal(t, []int{3, 5}, removed)
	assert.Equal(t, 3, len(l.slots))

	val, found := l.Get(0)
	assert.True(t, found)
	assert.Equal(t, "a", val)

	val, found = l.Get(1)
	assert.True(t, found)
	assert.Equal(t, "d", val)

	val, found = l.Get(2)
	assert.True(t, found)
	assert.Equal(t, "f", val)
}

func TestFreeList_Iter(t *testing.T) {
	l := NewFreeList[string]()
	assert.Equal(t, 0, l.Add("a"))
	assert.Equal(t, 1, l.Add("b"))
	assert.Equal(t, 2, l.Add("c"))

	for idx, item := range l.Iter() {
		switch idx {
		case 0:
			assert.Equal(t, "a", item)
		case 1:
			assert.Equal(t, "b", item)
		case 2:
			assert.Equal(t, "c", item)
		default:
			assert.Failf(t, "invalid", "idx: %v", idx)
		}
	}

	// remove one item in the middle
	assert.True(t, l.Remove(1))
	for idx, item := range l.Iter() {
		switch idx {
		case 0:
			assert.Equal(t, "a", item)
		case 2:
			assert.Equal(t, "c", item)
		default:
			assert.Failf(t, "invalid", "idx: %v", idx)
		}
	}
}
