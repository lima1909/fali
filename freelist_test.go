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
