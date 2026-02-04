package main

import (
	"reflect"
	"sync"
)

// GetFieldValue helper function to get the value for a given type: T (mostly a struct)
type FieldValueFn[T any] = func(*T) any

type IndexList[T any] struct {
	list          FreeList[T]
	allIDs        BitSet[uint32]
	fieldIndexMap FieldIndexMap[T, uint32]

	lock sync.RWMutex
}

func NewIndexList[T any]() *IndexList[T] {
	return &IndexList[T]{
		list:          NewFreeList[T](),
		allIDs:        BitSet[uint32]{data: make([]uint64, 0)},
		fieldIndexMap: NewFieldIndexMap[T, uint32](),
	}
}

func (l *IndexList[T]) Add(item T) int {
	l.lock.Lock()
	defer l.lock.Unlock()

	row := l.list.Add(item)
	l.allIDs.Set(uint32(row))

	for name, fieldIndex := range l.fieldIndexMap {
		val := fieldIndex.fieldFn(&item)

		// safe the type of val to validate it before executing the Query
		if fieldIndex.fieldFnResultType == nil {
			fieldIndex.fieldFnResultType = reflect.TypeOf(val)
			l.fieldIndexMap[name] = fieldIndex
		}

		fieldIndex.index.Set(val, uint32(row))
	}

	return row
}

func (l *IndexList[T]) Query(query Query[uint32]) (int, error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	bs, _, err := query(l.fieldIndexMap.IndexByName, &l.allIDs)
	if err != nil {
		return 0, err
	}

	count := 0
	bs.Values(func(v uint32) bool {
		count++
		return true
	})

	return count, nil
}
