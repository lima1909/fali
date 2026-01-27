package main

import (
	"math/bits"
)

type Value interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type BitSet[V Value] struct {
	data []uint64
}

// NewBitSet creates a new BitSet
func NewBitSet[V Value]() *BitSet[V] {
	return NewBitSetWithCapacity[V](0)
}

// NewBitSetFrom creates a new BitSet with given values
func NewBitSetFrom[V Value](values ...V) *BitSet[V] {
	b := NewBitSetWithCapacity[V](len(values))
	for _, v := range values {
		b.Set(v)
	}
	return b
}

// NewBitSetWithCapacity creates a new BitSet with starting capacity
func NewBitSetWithCapacity[V Value](size int) *BitSet[V] {
	return &BitSet[V]{data: make([]uint64, size)}
}

// Set inserts or updates the key in the BitSet
func (b *BitSet[V]) Set(value V) {
	l := V(len(b.data))
	idx := value >> 6

	if idx >= l {
		// resize the slice if necessary
		newCap := max(idx+1, l*2)
		newData := make([]uint64, newCap)
		copy(newData, b.data)
		b.data = newData
	}

	b.data[idx] |= (1 << (value & 63))

	// i>>6 is equals i/64 but faster
	// i&63 is the same: i%64, but faster
}

// Delete removes the key from the BitSet. Clear the bit value to 0.
func (b *BitSet[V]) UnSet(value V) bool {
	if int(value)>>6 >= len(b.data) {
		return false
	}

	b.data[value>>6] &^= (1 << (value & 63))
	return true
}

// Contains check, is the value saved in the BitSet
func (b *BitSet[V]) Contains(value V) bool {
	if int(value)>>6 >= len(b.data) {
		return false
	}

	return (b.data[value>>6] & (1 << (value & 63))) != 0
}

// Min return the min value where an Bit is set
// [1, 3, 100] => 1
// if no max found, return -1
func (b *BitSet[V]) Min() int {
	bl := len(b.data)
	bd := b.data

	for i := range bl {
		w := bd[i]
		if w != 0 {
			// bits.TrailingZeros64 returns the number of zero bits
			// before the first set bit (the "1").
			// Example: w = ...1000 (binary) -> TrailingZeros64 returns 3.
			// The index of that bit is exactly 3.
			return (i << 6) + bits.TrailingZeros64(w)
		}
	}

	return -1
}

// Max return the max value where an Bit is set
// [1, 3, 100] => 100
// if no max found, return -1
func (b *BitSet[V]) Max() int {
	bl := len(b.data)
	bd := b.data

	for i := bl - 1; i >= 0; i-- {
		w := bd[i]
		if w != 0 {
			// bits.Len64 returns the minimum bits to represent w.
			// Example: w = 0...0101 (binary) -> Len64 returns 3.
			// The index of that bit is 3 - 1 = 2.
			return (i << 6) + (bits.Len64(w) - 1)
		}
	}

	return -1
}

// MaxSetIndex return the max index where an Bit is set
func (b *BitSet[V]) MaxSetIndex() int {
	bl := len(b.data)
	bd := b.data

	for i := bl - 1; i >= 0; i-- {
		if bd[i] != 0 {
			return i
		}
	}

	return -1
}

// Counts how many values are in the BitSet, bits are set.
func (b *BitSet[V]) Count() int {
	bl := len(b.data)
	bd := b.data
	count := 0

	for i := range bl {
		count += bits.OnesCount64(bd[i])
	}

	return count
}

// Len returns the len of the bit slice
func (b *BitSet[V]) Len() int {
	return len(b.data)
}

// how many bytes is using
func (b *BitSet[V]) usedBytes() int {
	return 24 + (len(b.data) * 8)
}

// Copy copy the complete BitSet.
func (b *BitSet[V]) Copy() *BitSet[V] {
	target := make([]uint64, len(b.data))
	copy(target, b.data)
	return &BitSet[V]{data: target}
}

// And is the logical AND of two BitSet
// In this BitSet is the result, this means the values will be overwritten!
func (b *BitSet[V]) And(other *BitSet[V]) {
	bd := b.data
	od := other.data

	l := min(len(b.data), len(other.data))
	for i := range l {
		bd[i] &= od[i]
	}

	// remove the tail
	// This updates the 'len', but the 'cap' (memory) remains.
	b.data = bd[:l]
}

// Or is the logical OR of two BitSet
func (b *BitSet[V]) Or(other *BitSet[V]) {
	bd := b.data
	od := other.data
	ol := len(od)
	bl := len(bd)

	// other is longer - we must grow and handle the tail
	if ol > bl {
		if cap(bd) < ol {
			// new allocation
			newData := make([]uint64, ol)
			copy(newData, bd)
			b.data = newData
		} else {
			// reuse capacity
			b.data = bd[:ol]
		}

		// only OR the overlapping part
		for i := range bl {
			b.data[i] |= od[i]
		}
		// Directly COPY the tail of 'other' into 'b'
		// This overwrites any "ghost bits" in the reused capacity
		copy(b.data[bl:], od[bl:])
		return
	}

	// b is longer or equal - no growth needed
	for i := range ol {
		b.data[i] |= od[i]
	}
}

// XOr is the logical XOR of two BitSet
func (b *BitSet[V]) Xor(other *BitSet[V]) {
	bd := b.data
	od := other.data
	ol := len(od)
	bl := len(bd)

	// resize b
	if ol > bl {
		// Grow b to match other
		if cap(bd) < ol {
			newData := make([]uint64, ol)
			copy(newData, bd)
			b.data = newData
		} else {
			// trim to the len of ohter
			b.data = bd[:ol]
		}

		// execute xor for ol > bl
		for i := range bl {
			b.data[i] ^= od[i]
		}
		copy(b.data[bl:], od[bl:])
	} else {
		// execute xor for ol <= bl
		for i := range ol {
			b.data[i] ^= od[i]
		}
	}
}

// AndNot removes all elements from the current set that exist in another set.
// Known as "Bit Clear" or "Set Difference"
//
// Example: [1, 2, 110, 2345] AndNot [2, 110] => [1, 2345]
func (b *BitSet[V]) AndNot(other *BitSet[V]) {
	od := other.data

	// we only care about the intersection.
	// If 'other' is longer, we ignore its tail (0 &^ 1 = 0).
	// If 'b' is longer, its tail stays the same (1 &^ 0 = 1).
	limit := min(len(b.data), len(od))

	// perform the Bit Clear operation
	for i := range limit {
		b.data[i] &^= od[i]
	}
}

// Shrink trims the bitset to ensure that len(b.data) always points to the last truly useful word.
//
// Operation	Can Grow?	Can Shrink?
// OR	        Yes	        No
// XOR	        Yes     	Yes
// AND	        No      	Yes
// AND NOT      No	        Yes
func (b *BitSet[V]) Shrink() {
	bd := b.data
	bl := len(bd)

	for i := bl - 1; i >= 0; i-- {
		if bd[i] != 0 {
			b.data = bd[:i+1]
			return
		}
	}

	// all data are equals 0
	b.data = bd[:0]
}

// ToSlice create a new slice which contains all saved values
func (b *BitSet[V]) ToSlice() []int {
	res := make([]int, 0, b.Count())
	bd := b.data
	l := len(bd)

	for i := range l {
		w := bd[i]
		for w != 0 {
			t := bits.TrailingZeros64(w)
			res = append(res, (i<<6)+t)
			w &^= (1 << uint(t)) // Clear the bit we just found
		}
	}
	return res
}
