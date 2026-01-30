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
	bd := b.data

	for i, w := range bd {
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
	bd := b.data
	count := 0

	for _, w := range bd {
		count += bits.OnesCount64(w)
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

	l := min(len(bd), len(od))

	// Bounds Check Elimination
	a := bd[:l]
	o := od[:l]

	// process 4 words (256 bits) per iteration.
	i := 0
	for ; i <= l-4; i += 4 {
		a[i] &= o[i]
		a[i+1] &= o[i+1]
		a[i+2] &= o[i+2]
		a[i+3] &= o[i+3]
	}

	// handle remaining elements (the tail)
	for ; i < l; i++ {
		a[i] &= o[i]
	}

	// remove the tail
	// This updates the 'len', but the 'cap' (memory) remains.
	b.data = a
}

// Or is the logical OR of two BitSet
func (b *BitSet[V]) Or(other *BitSet[V]) {
	od := other.data
	ol := len(od)

	if len(b.data) < ol {
		if cap(b.data) < ol {
			newD := make([]uint64, ol)
			copy(newD, b.data)
			b.data = newD
		} else {
			b.data = b.data[:ol]
		}
	}

	// BCE (Bounds Check Elimination)
	// are safe to access up to index 'ol-1'.
	dst := b.data[:ol]
	src := od

	// unrolled Loop for the overlapping part
	i := 0
	for ; i <= ol-4; i += 4 {
		dst[i] |= src[i]
		dst[i+1] |= src[i+1]
		dst[i+2] |= src[i+2]
		dst[i+3] |= src[i+3]
	}

	// clean up the remainder
	for ; i < ol; i++ {
		dst[i] |= src[i]
	}

	// if len(b.data) was originally > ol, the rest of b.data
	// stays exactly as it was, which is correct for an OR operation.
}

// XOr is the logical XOR of two BitSet
func (b *BitSet[V]) Xor(other *BitSet[V]) {
	od := other.data
	ol := len(od)
	bl := len(b.data)

	common := ol
	if bl < ol {
		if cap(b.data) < ol {
			newD := make([]uint64, ol)
			copy(newD, b.data)
			b.data = newD
		} else {
			b.data = b.data[:ol]
		}
		common = bl
	}

	// setup BCE (Bounds Check Elimination)
	// only XOR up to the 'common' length (where both have data)
	dst := b.data[:common]
	src := od[:common]

	// unrolled Loop for the overlapping part
	i := 0
	for ; i <= common-4; i += 4 {
		dst[i] ^= src[i]
		dst[i+1] ^= src[i+1]
		dst[i+2] ^= src[i+2]
		dst[i+3] ^= src[i+3]
	}

	for ; i < common; i++ {
		dst[i] ^= src[i]
	}

	// handle the Tail
	// if other was longer than b, the tail of other should be copied into b.
	// (Because 0 XOR Value = Value)
	if ol > bl {
		copy(b.data[bl:], od[bl:])
	}
	// Note: If b was longer than other, the tail of b remains as is.
	// (Because Value XOR 0 = Value)
}

// AndNot removes all elements from the current set that exist in another set.
// Known as "Bit Clear" or "Set Difference"
//
// Example: [1, 2, 110, 2345] AndNot [2, 110] => [1, 2345]
func (b *BitSet[V]) AndNot(other *BitSet[V]) {
	bd := b.data
	od := other.data

	l := min(len(bd), len(od))
	// BCE (Bounds Check Elimination)
	a := bd[:l]
	o := od[:l]

	// unrolling (Process 4 words per iteration)
	i := 0
	for ; i <= l-4; i += 4 {
		a[i] &^= o[i]
		a[i+1] &^= o[i+1]
		a[i+2] &^= o[i+2]
		a[i+3] &^= o[i+3]
	}

	// handle remaining elements
	for ; i < l; i++ {
		a[i] &^= o[i]
	}

	// Note: b.data's tail beyond 'l' is left untouched.
	// 1 &^ 0 = 1, so the bits naturally stay set.
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

	// start from the end
	i := len(bd) - 1
	for i >= 0 && bd[i] == 0 {
		i--
	}

	// update length once
	b.data = bd[:i+1]
}

// Values iterate over the complete BitSet and call the yield function, for every value
func (b *BitSet[V]) Values(yield func(V) bool) {
	bd := b.data
	for i, w := range bd {
		for w != 0 {
			t := bits.TrailingZeros64(w)
			val := (i << 6) + t
			if !yield(V(val)) {
				return
			}
			w &= (w - 1)
		}
	}
}

// ToSlice create a new slice which contains all saved values
func (b *BitSet[V]) ToSlice() []V {
	res := make([]V, 0, b.Count())
	b.Values(func(v V) bool {
		res = append(res, v)
		return true
	})
	return res
}
