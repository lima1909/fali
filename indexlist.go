package main

import (
	"reflect"
	"sort"
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

// Get return a Item for a given index, if an Item exist for this index.
// Otherwise is found = false.
func (l *IndexList[T]) Get(index int) (t T, found bool) {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.list.Get(index)
}

// Add add the given Item to the list,
// there is NO check, for existing this Item in the list
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

func (l *IndexList[T]) Remove(index int) (t T, removed bool) {
	item, found := l.Get(index)
	if !found {
		return item, found
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	removed = l.list.Remove(index)
	l.allIDs.UnSet(uint32(index))

	for _, fieldIndex := range l.fieldIndexMap {
		val := fieldIndex.fieldFn(&item)
		fieldIndex.index.UnSet(val, uint32(index))
	}

	return item, removed
}

// Query execute the given Query.
func (l *IndexList[T]) Query(query Query[uint32]) (QueryResult[T], error) {
	l.lock.RLock()
	defer l.lock.RUnlock()

	bs, _, err := query(l.fieldIndexMap.IndexByName, &l.allIDs)
	if err != nil {
		return QueryResult[T]{}, err
	}

	return QueryResult[T]{bitSet: *bs, list: l}, nil
}

// Count the Items, which in this list exist
func (l *IndexList[T]) Count() int {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.allIDs.Count()
}

type QueryResult[T any] struct {
	bitSet BitSet[uint32]
	list   *IndexList[T]
}

func (q *QueryResult[T]) Count() int        { return q.bitSet.Count() }
func (q *QueryResult[T]) Empty() bool       { return q.bitSet.Count() == 0 }
func (q *QueryResult[T]) Indices() []uint32 { return q.bitSet.ToSlice() }

func (q *QueryResult[T]) Values() []T {
	list := make([]T, 0, q.bitSet.Count())

	q.list.lock.RLock()
	q.bitSet.Values(func(r uint32) bool {
		// get from the FreeList without lock
		o, _ := q.list.list.Get(int(r))
		list = append(list, o)

		return true
	})
	q.list.lock.RUnlock()

	return list
}

func (q *QueryResult[T]) Sort(less func(*T, *T) bool) []T {
	list := q.Values()
	sort.Slice(list, func(i, j int) bool { return less(&list[i], &list[j]) })
	return list
}

func (q *QueryResult[T]) Remove() int {
	count := 0
	q.bitSet.Values(func(r uint32) bool {
		if _, removed := q.list.Remove(int(r)); removed {
			count++
		}
		return true
	})

	return count
}
