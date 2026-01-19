package main

import (
	"fmt"
	"math/rand"
	"time"
)

// https://en.wikipedia.org/wiki/Skip_list

const (
	maxLevel = 16 // supports up to ~4.3 million elements
	p        = 0.25
)

type node[V any] struct {
	key   uint32 // maybe uint16 is enough (65_536 elements)
	value V
	level byte
	next  [maxLevel]*node[V]
}

func (n node[V]) String() string {
	return fmt.Sprintf("{K: %d, L: %d}", n.key, n.level)
}

type SkipList[V any] struct {
	head  *node[V]
	level byte
}

// randomLevel generates a random height
func randomLevel() byte {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	lvl := byte(1)
	for lvl < maxLevel && rnd.Float64() < p {
		lvl++
	}
	return lvl
}

// New creates a new SkipList
func New[V any]() *SkipList[V] {
	return &SkipList[V]{
		head:  &node[V]{level: maxLevel},
		level: 1,
	}
}

// Get returns value and whether it exists
func (sl *SkipList[V]) Get(key uint32) (V, bool) {
	x := sl.head
	i := sl.level

	for i > 0 {
		i--
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
func (sl *SkipList[V]) Put(key uint32, value V) {
	update := [maxLevel]*node[V]{}
	x := sl.head
	i := sl.level

	for i > 0 {
		i--
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
		return
	}

	lvl := randomLevel()
	if lvl > sl.level {
		for i := sl.level; i < lvl; i++ {
			update[i] = sl.head
		}
		sl.level = lvl
	}

	// create new insert node
	n := &node[V]{
		key:   key,
		value: value,
		level: lvl,
	}

	for i := range lvl {
		n.next[i] = update[i].next[i]
		update[i].next[i] = n
	}
}

// Delete removes the value for a given key
func (sl *SkipList[V]) Delete(key uint32) bool {
	update := [maxLevel]*node[V]{}
	x := sl.head
	i := sl.level

	for i > 0 {
		i--
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

func (sl *SkipList[V]) Print() {
	if sl.head == nil || sl.head.next[0] == nil {
		return
	}

	fmt.Printf("[L: %d] :: ", sl.level)

	x := sl.head
	for x != nil {
		// ignore head for printing
		if x == sl.head {
			x = x.next[0]
			continue
		}

		fmt.Printf("%s, ", x)
		x = x.next[0]
	}
	fmt.Printf("\n")
}

func (sl *SkipList[V]) PrintLevels() {
	x := sl.head
	for x != nil {

		if x.next[0] != nil {
			fmt.Printf("{K: %d :: ", x.next[0].key)
			// for i := x.level - 1; i >= 0; i-- {
			i := sl.level
			for i > 0 {
				i--
				if x.next[i] != nil {
					fmt.Printf("%d, ", x.next[i].key)
				}
			}
			fmt.Printf("} ")
		}
		x = x.next[0]
	}
	fmt.Printf("\n")
}

func main() {
	sl := New[int]()
	sl.Print()
	sl.Put(2, 2)
	sl.Print()
	sl.Put(1, 1)
	sl.Print()
	sl.Put(3, 3)
	sl.Print()
	sl.Put(4, 4)
	sl.Print()
	sl.Put(5, 5)
	sl.Print()

	sl.PrintLevels()

	fmt.Println(sl.Get(2))
	fmt.Println(sl.Get(5))

}
