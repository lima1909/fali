package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const ()

func BenchmarkBitSetContains(b *testing.B) {
	bs := NewBitSet[uint32]()
	for i := 1; i <= count; i++ {
		bs.Set(uint32(i))
	}
	b.ResetTimer()

	for b.Loop() {
		assert.True(b, bs.Contains(found_val))
	}
}

func BenchmarkBitSetCount(b *testing.B) {
	bs := NewBitSet[uint32]()
	for i := 1; i <= count; i++ {
		bs.Set(uint32(i))
	}
	b.ResetTimer()

	for b.Loop() {
		assert.Equal(b, count, bs.Count())
	}
}

// func BenchmarkBitSetCountFast(b *testing.B) {
// 	bs := NewBitSet()
// 	for i := 1; i <= values; i++ {
// 		bs.Set(i)
// 	}
// 	b.ResetTimer()
//
// 	for b.Loop() {
// 		assert.Equal(b, values, bs.CountFast())
// 	}
// }
