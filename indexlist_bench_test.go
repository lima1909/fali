package main

import (
	_ "embed"

	"fmt"
	"math/rand/v2"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//go:embed testdata/names.txt
var names_txt string

func BenchmarkQueryStr(b *testing.B) {
	type person struct {
		Name string
		Age  int
	}

	minV := 10
	maxV := 100
	names := strings.Split(names_txt, "\n")

	start := time.Now()

	// Sprintf is expensive, so we only test with 250_000 datasets
	ds := 3_000_000
	il := NewIndexList[person]()
	err := il.CreateIndex("name", NewSortedIndex(FromName[person, string]("Name")))
	assert.NoError(b, err)
	err = il.CreateIndex("age", NewSortedIndex(FromName[person, int]("Age")))
	assert.NoError(b, err)

	n := 0
	for i := 1; i <= ds; i++ {
		if n%6779 == 0 {
			n = 0
		}
		n++

		il.Insert(person{
			Name: names[n],
			Age:  minV + rand.IntN(maxV-minV+1),
		})
	}
	fmt.Printf("- Count: %d, Time: %s\n", il.Count(), time.Since(start))
	b.ResetTimer()

	count := 0
	for b.Loop() {
		qr, err := il.QueryStr(
			`name = "Jule" or name = "Magan" or age > int(80)`,
		)
		assert.NoError(b, err)
		count = max(count, qr.Count())
	}

	fmt.Printf("- Max count: %d\n\n", count)
}
