package main

// Slot holds the data or the pointer to the next free space
type slot[T any] struct {
	value    T
	nextFree int  // If Occupied=false, this points to the next available slot
	occupied bool // Simple flag to know if this is data or a free link
}

type FreeList[T any] struct {
	slots    []slot[T]
	freeHead int // Index of the first free slot (-1 if none)
}

func NewFreeList[T any]() *FreeList[T] {
	return &FreeList[T]{
		slots:    make([]slot[T], 0),
		freeHead: -1, // -1 means "No free slots, append new ones"
	}
}

func (l *FreeList[T]) Add(item T) int {
	// no free slots in the list, append to the end
	if l.freeHead == -1 {
		idx := len(l.slots)
		l.slots = append(l.slots, slot[T]{
			value:    item,
			occupied: true,
			nextFree: -1,
		})
		return idx
	}

	idx := l.freeHead
	l.freeHead = l.slots[idx].nextFree
	l.slots[idx] = slot[T]{
		value:    item,
		occupied: true,
		nextFree: -1,
	}

	return idx
}

func (l *FreeList[T]) Remove(index int) bool {
	if index < 0 || index >= len(l.slots) || !l.slots[index].occupied {
		return false
	}

	// clear the value to prevent memory leaks
	var null T
	l.slots[index].value = null
	l.slots[index].occupied = false

	// make this slot point to the current head
	l.slots[index].nextFree = l.freeHead
	// make this slot the new head
	l.freeHead = index

	return true
}

func (l *FreeList[T]) Get(index int) (T, bool) {
	if index < 0 || index >= len(l.slots) {
		var null T
		return null, false
	}

	slot := l.slots[index]
	if !slot.occupied {
		var null T
		return null, false
	}
	return slot.value, true
}
