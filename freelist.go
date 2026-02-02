package main

// Slot holds the data or the pointer to the next free space
type slot[T any] struct {
	value    T
	nextFree int  // If Occupied=false, this points to the next available slot
	occupied bool // Simple flag to know if this is data or a free link
}

// FreeList don't delete an Item, instead mark it as not occupied.
// With one of the Compact Methods, you can remove thes palceholders and make the list smaller.
type FreeList[T any] struct {
	slots    []slot[T]
	freeHead int // Index of the first free slot (-1 if none)
}

func NewFreeList[T any]() FreeList[T] {
	return FreeList[T]{
		slots:    make([]slot[T], 0),
		freeHead: -1, // -1 means "No free slots, append new ones"
	}
}

// Add an Item to the end of the List or use a free slot, to add this item
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

// Remove mark the Item on the given index as deleted.
// index must be >=0 and < len(slots), otherwise return Remove false and do nothing.
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

// Get the Item on the given index, or the zero value and false, if it not exist.
// index must be >=0 and < len(slots), otherwise return Get zero value and false and do nothing.
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

// CompactUnstable removes not used slots. Unstable means, the Indices breaks.
func (l *FreeList[T]) CompactUnstable() {
	keep := 0
	slots := l.slots

	for _, s := range slots {
		if s.occupied {
			slots[keep] = s
			keep++
		}
	}

	l.slots = slots[:keep]
	l.freeHead = -1
}

// CompactLinear removes not used slote.
// If an Index has changed, yout get this Info with the Callback: onMove
func (l *FreeList[T]) CompactLinear(onMove func(oldIndex, newIndex int)) {
	var null T
	keep := 0
	slots := l.slots

	for i, s := range slots {
		if s.occupied {
			// If the read and write pointers are different, we need to move the data
			if i != keep {
				slots[keep] = s
				onMove(i, keep)

				// clear the old slot to prevent memory leaks
				slots[i] = slot[T]{value: null, occupied: false, nextFree: -1}
			}
			keep++
		}
	}

	l.slots = slots[:keep]
	l.freeHead = -1
}
