package main

import (
	"math/rand"
	"time"
)

// A SkipList is a data structure that allows for fast search, insertion, and deletion within a sorted list.
// It acts as an alternative to balanced binary search trees.
//
// Think of a SkipList as a standard Sorted Linked List but with "express lanes."
//
// https://en.wikipedia.org/wiki/Skip_list
//

const (
	maxLevel   = 16 // supports up to ~4.3 million elements
	population = 0.25
)

type node[V any] struct {
	key   uint32 // maybe uint16 is enough (65_536 elements)
	value V
	level byte
	next  [maxLevel]*node[V]
}

type SkipList[V any] struct {
	head  *node[V]
	level byte

	rnd *rand.Rand
}

// randomLevel generates a random height (level)
func (sl *SkipList[V]) randomLevel() byte {
	lvl := byte(1)
	for lvl < maxLevel && sl.rnd.Float64() < population {
		lvl++
	}
	return lvl
}

// NewSkipList creates a new SkipList
func NewSkipList[V any]() *SkipList[V] {
	return &SkipList[V]{
		head:  &node[V]{level: maxLevel},
		level: 1,
		rnd:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Get returns value and whether it exists
func (sl *SkipList[V]) Get(key uint32) (V, bool) {
	x := sl.head
	for i := int(sl.level) - 1; i >= 0; i-- {
		for x.next[i] != nil && x.next[i].key < key {
			x = x.next[i]
		}
	}

	x = x.next[0]
	if x != nil && x.key == key {
		// key found
		return x.value, true
	}

	var zeroVal V
	return zeroVal, false
}

// Put inserts or updates a key, with the given value
// If the set did not previously contain this value, true is returned, otherwise false.
func (sl *SkipList[V]) Put(key uint32, value V) bool {
	update := [maxLevel]*node[V]{}
	x := sl.head
	for i := int(sl.level) - 1; i >= 0; i-- {
		// move forward while next node's key < insertion key
		for x.next[i] != nil && x.next[i].key < key {
			x = x.next[i]
		}
		// save the last node visited at this level
		update[i] = x
	}

	// node found: UPDATE
	x = x.next[0]
	if x != nil && x.key == key {
		x.value = value
		return false
	}

	lvl := sl.randomLevel()
	if lvl > sl.level {
		for i := sl.level; i < lvl; i++ {
			update[i] = sl.head
		}
		sl.level = lvl
	}

	// create new insert node
	n := &node[V]{key: key, value: value, level: lvl}
	for i := range lvl {
		n.next[i] = update[i].next[i]
		update[i].next[i] = n
	}

	return true
}

// Delete removes the value for a given key
// If the key was not found: false, otherwise true, if the key was deleted.
func (sl *SkipList[V]) Delete(key uint32) bool {
	update := [maxLevel]*node[V]{}
	x := sl.head
	for i := int(sl.level) - 1; i >= 0; i-- {
		for x.next[i] != nil && x.next[i].key < key {
			x = x.next[i]
		}
		update[i] = x
	}

	x = x.next[0]
	if x == nil || x.key != key {
		// not found, no value deleted
		return false
	}

	for i := byte(0); i < sl.level; i++ {
		if update[i].next[i] != x {
			break
		}
		update[i].next[i] = x.next[i]
	}

	for sl.level > 1 && sl.head.next[sl.level-1] == nil {
		sl.level--
	}

	return true
}

// Range returns all values for keys in the range [from, to] inclusive
func (sl *SkipList[V]) Range(from, to uint32) []V {
	if from > to {
		return nil
	}

	// find the first node >= from
	x := sl.head
	for i := int(sl.level) - 1; i >= 0; i-- {
		for x.next[i] != nil && x.next[i].key < from {
			x = x.next[i]
		}
	}

	// move to the actual first node at level 0
	x = x.next[0]
	if x == nil || x.key > to {
		return nil
	}

	// collect all nodes until we exceed 'to'
	result := make([]V, 0, 16)
	for x != nil && x.key <= to {
		result = append(result, x.value)
		x = x.next[0] // Always stay on the ground floor (Level 0)
	}

	return result
}
