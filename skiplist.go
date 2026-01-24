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

type VisitFn[V any] func(key uint32, val V) bool

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
//
//go:inline
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

// Put inserts or updates a key with the given value.
// Returns true if a new node was inserted, false if an existing key was updated.
func (sl *SkipList[V]) Put(key uint32, value V) bool {
	update := [maxLevel]*node[V]{}
	x := sl.head

	// search for the position and fill the 'update' array
	for i := int(sl.level) - 1; i >= 0; i-- {
		// move forward while next node's key < insertion key
		for next := x.next[i]; next != nil && next.key < key; next = x.next[i] {
			x = next
		}
		// save the last node visited at this level
		update[i] = x
	}

	// check if the key already exists
	x = x.next[0]
	if x != nil && x.key == key {
		x.value = value // update existing value
		return false    // not a new insertion
	}

	// key does not exist, prepare new node level
	lvl := sl.randomLevel()

	// if the new level is higher than current, initialize 'update' for the gap
	if lvl > sl.level {
		for i := sl.level; i < lvl; i++ {
			update[i] = sl.head
		}
	}

	// create and link the new node
	n := &node[V]{key: key, value: value, level: lvl}

	for i := range lvl {
		n.next[i] = update[i].next[i]
		update[i].next[i] = n
	}

	// update global level
	if lvl > sl.level {
		sl.level = lvl
	}

	return true
}

// Delete removes the value for a given key
// If the key was not found: false, otherwise true, if the key was deleted.
func (sl *SkipList[V]) Delete(key uint32) bool {
	update := [maxLevel]*node[V]{}
	x := sl.head
	for i := int(sl.level) - 1; i >= 0; i-- {
		for next := x.next[i]; next != nil && next.key < key; next = x.next[i] {
			x = next
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

// Traverse over the complete Skiplist and calling the visitor
// the return value false means, not to the end, otherwise true
func (sl *SkipList[V]) Traverse(visit VisitFn[V]) bool {
	for next := sl.head.next[0]; next != nil; next = next.next[0] {
		// break if false (simulate yield)
		if !visit(next.key, next.value) {
			// reads not to the end
			return false
		}
	}

	return true
}

// Range traverse 'from' until 'to' over Skiplist and calling the visitor
func (sl *SkipList[V]) Range(from, to uint32, visit VisitFn[V]) {
	if from > to {
		return
	}

	// find the first node >= from
	x := sl.head
	for i := int(sl.level) - 1; i >= 0; i-- {
		for next := x.next[i]; next != nil && next.key < from; next = x.next[i] {
			x = next
		}
	}

	// move to the actual first node at level 0
	x = x.next[0]
	if x == nil || x.key > to {
		return
	}

	// collect all nodes until we exceed 'to'
	for x != nil && x.key <= to {
		if !visit(x.key, x.value) {
			return
		}
		x = x.next[0] // Always stay on the ground floor (Level 0)
	}
}

// Get returns value and whether it exists
func (sl *SkipList[V]) Get(key uint32) (V, bool) {
	x := sl.head
	for i := int(sl.level) - 1; i >= 0; i-- {
		for next := x.next[i]; next != nil && next.key < key; next = x.next[i] {
			x = next
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

// Min returns the value associated with the smallest key in the list.
func (sl *SkipList[V]) Min() (V, bool) {
	// The first node on the bottom level (level 0) is the minimum
	first := sl.head.next[0]
	if first == nil {
		var zero V
		return zero, false
	}

	return first.value, true
}

// Max returns the value associated with the largest key in the list O(log n).
func (sl *SkipList[V]) Max() (V, bool) {
	x := sl.head
	// start at the highest lane and jump as far right as possible
	for i := int(sl.level) - 1; i >= 0; i-- {
		for x.next[i] != nil {
			x = x.next[i]
		}
	}

	// list is empty
	if x == sl.head {
		var zero V
		return zero, false
	}

	return x.value, true
}
