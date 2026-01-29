package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func BenchmarkBitSetAnd(b *testing.B) {
	bs1 := NewBitSet[uint32]()
	for i := 1; i <= count; i++ {
		if i%3 == 0 {
			bs1.Set(uint32(i))
		}
	}
	bs2 := NewBitSet[uint32]()
	for i := 1; i <= count; i++ {
		if i%6 == 0 {
			bs2.Set(uint32(i))
		}
	}
	b.ResetTimer()

	for b.Loop() {
		r := bs2.Copy()
		r.And(bs1)
		assert.Equal(b, 500_000, r.Count())
	}
}

func BenchmarkBitSetToSlice(b *testing.B) {
	bs := NewBitSet[uint32]()
	for i := 1; i <= count; i++ {
		bs.Set(uint32(i))
	}
	b.ResetTimer()

	for b.Loop() {
		assert.Equal(b, count, len(bs.ToSlice()))
	}
}

func BenchmarkBitSetValuesIter(b *testing.B) {
	bs := NewBitSet[uint32]()
	for i := 1; i <= count; i++ {
		bs.Set(uint32(i))
	}
	b.ResetTimer()

	for b.Loop() {
		c := 0
		bs.Values(func(v uint32) bool {
			_ = v
			c += 1
			return true

		})

		assert.Equal(b, count, c)

	}
}
