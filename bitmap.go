package main

//
// import (
// 	"math/bits"
// )
//
// // Would you like me to show you how to implement a "Sparse BitSet" (Roaring Bitmap style) for cases where you have a few bits set very far apart (e.g., bit 1 and bit 1,000,000)?
// type bitMap[V any] struct {
// 	bits   *BitSet
// 	values []V
// }
//
// func newbitMap[V any]() *bitMap[V] {
// 	return &bitMap[V]{
// 		bits:   NewBitSet(), // Using the dynamic BitSet we built
// 		values: make([]V, 0),
// 	}
// }
//
// // Get retrieves the value for a uint32 key.
// // Performance: O(N/64) to calculate Rank, then O(1) access.
// func (bm *bitMap[V]) Get(key uint32) (V, bool) {
// 	if !bm.bits.Contains(int(key)) {
// 		var zero V
// 		return zero, false
// 	}
//
// 	// The magic: The value's index in the dense slice is
// 	// exactly how many bits are set before it.
// 	index := bm.bits.rank(int(key))
// 	return bm.values[index], true
// }
//
// // Put inserts or updates a value.
// // Note: Inserting a NEW key is O(N) because the dense slice must shift.
// // Updating an EXISTING key is O(1).
// func (bm *bitMap[V]) Put(key uint32, val V) {
// 	k := int(key)
// 	if bm.bits.Contains(k) {
// 		// Update existing
// 		index := bm.bits.rank(k)
// 		bm.values[index] = val
// 		return
// 	}
//
// 	// Insert new
// 	index := bm.bits.rank(k)
// 	bm.bits.Set(k)
//
// 	// Standard Go slice insertion:
// 	// values = append(values[:index], append([]V{val}, values[index:]...)...)
// 	// But let's do it faster/cleaner:
// 	bm.values = append(bm.values, val) // Grow capacity
// 	if index < len(bm.values)-1 {
// 		copy(bm.values[index+1:], bm.values[index:])
// 		bm.values[index] = val
// 	}
// }
//
// // rank counts, in wich []data are bits
// // Example: if you have bits set at positions 3, 50, and 500, your values slice only has 3 items.
// func (b *BitSet) rank(i int) int {
// 	targetWord := i >> 6
// 	targetBit := uint(i) & 63
// 	rank := 0
//
// 	// 1. Popcount full words (Very fast on modern CPUs)
// 	for j := 0; j < targetWord && j < len(b.data); j++ {
// 		rank += bits.OnesCount64(b.data[j])
// 	}
//
// 	// 2. Popcount the partial word up to the target bit
// 	if targetWord < len(b.data) {
// 		mask := (uint64(1) << targetBit) - 1
// 		rank += bits.OnesCount64(b.data[targetWord] & mask)
// 	}
//
// 	return rank
// }
