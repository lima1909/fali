package main

import (
	"sync"
)

// GetFieldValue helper function to get the value for a given type: T (mostly a struct)
type FieldValueFn[T any] = func(*T) any

type IndexList[T any] struct {
	list         FreeList[T]
	allIDs       BitSet[uint32]
	fieldIndex   FieldIndex[uint32]
	fieldValueFn map[string]FieldValueFn[T] //TODO: maybe, we can put the two maps together

	lock sync.RWMutex
}

func NewIndexList[T any]() *IndexList[T] {
	return &IndexList[T]{
		list:         NewFreeList[T](),
		allIDs:       BitSet[uint32]{data: make([]uint64, 0)},
		fieldIndex:   make(map[string]Index[uint32], 0),
		fieldValueFn: make(map[string]FieldValueFn[T], 0),
	}
}

func (l *IndexList[T]) Add(item T) int {
	l.lock.Lock()
	defer l.lock.Unlock()

	row := l.list.Add(item)
	l.allIDs.Set(uint32(row))

	for name, get := range l.fieldValueFn {
		val := get(&item)
		// fmt.Println("--", row, name, val)
		idx := l.fieldIndex[name]
		idx.Set(val, uint32(row))
	}

	return row
}

func (l *IndexList[T]) Query(q Query[uint32]) int {
	bs, _ := q(l.fieldIndex, &l.allIDs)
	count := 0
	bs.Values(func(v uint32) bool {
		count++
		return true
	})

	return count
}
