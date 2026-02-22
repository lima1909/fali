package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func BenchmarkQueryStr(b *testing.B) {
	start := time.Now()

	// Sprintf is expensive, so we only test with 250_000 datasets
	ds := 250_000
	il := NewIndexList[car]()
	err := il.CreateIndex("name", NewSortedIndex((*car).Name))
	assert.NoError(b, err)

	for i := 1; i <= ds; i++ {
		if i%2 == 0 {
			il.Insert(car{name: fmt.Sprintf("Mercedes %d", i), age: 5, isNew: true})
		} else if i%3 == 0 {
			il.Insert(car{name: "Dacia", age: 22})
			il.Insert(car{name: "Opel", age: 22})
		} else {
			il.Insert(car{name: fmt.Sprintf("Dacia %d", i), age: 22})
			il.Insert(car{name: fmt.Sprintf("Opel %d", i), age: 22})
		}
	}
	fmt.Println("--->>>", il.Count(), time.Since(start))
	b.ResetTimer()

	for b.Loop() {
		qr, err := il.QueryStr(`name = "Opel" or name = "Dacia"`)
		assert.NoError(b, err)
		assert.Equal(b, 83_334, qr.Count())
	}
}
