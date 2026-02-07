package main

import (
	"reflect"
	"sort"
	"sync"
)

// GetFieldValue helper function to get the value for a given type: T (mostly a struct)
type FieldValueFn[T any] = func(*T) any

// IndexList is a list (slice), which is extended by Indices for fast finding Items in the list.
type IndexList[T any] struct {
	list          FreeList[T]
	allIDs        BitSet[uint32]
	fieldIndexMap FieldIndexMap[T, uint32]

	lock sync.RWMutex
}

// NewIndexList create a new IndexList
func NewIndexList[T any]() *IndexList[T] {
	return &IndexList[T]{
		list:          NewFreeList[T](),
		allIDs:        BitSet[uint32]{data: make([]uint64, 0)},
		fieldIndexMap: NewFieldIndexMap[T, uint32](),
	}
}

// CreateIndex create a new Index:
//   - fieldName: a name for a field of the saved Item
//   - fieldGetFn: a function, which returns the value of an field
//   - Index: a impl of the Index interface
func (l *IndexList[T]) CreateIndex(fieldName string, fieldGetFn FieldGetFn[T], index Index[uint32]) {

	var t T

	fieldIndex := FieldIndex[T, uint32]{
		index:             index,
		fieldFn:           fieldGetFn,
		fieldFnResultType: reflect.TypeOf(fieldGetFn(&t)),
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	for idx, item := range l.list.Iter() {
		val := fieldIndex.fieldFn(&item)
		fieldIndex.index.Set(val, uint32(idx))
	}

	l.fieldIndexMap[fieldName] = fieldIndex
}

// Add add the given Item to the list,
// there is NO check, for existing this Item in the list
func (l *IndexList[T]) Add(item T) int {
	l.lock.Lock()
	defer l.lock.Unlock()

	idx := l.list.Add(item)
	l.allIDs.Set(uint32(idx))

	for _, fieldIndex := range l.fieldIndexMap {
		val := fieldIndex.fieldFn(&item)
		fieldIndex.index.Set(val, uint32(idx))
	}

	return idx
}

// Query execute the given Query.
func (l *IndexList[T]) Query(query Query[uint32]) (QueryResult[T], error) {
	l.lock.RLock()
	bs, _, err := query(l.fieldIndexMap.IndexByName, &l.allIDs)
	l.lock.RUnlock()

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

func (q *QueryResult[T]) Count() int  { return q.bitSet.Count() }
func (q *QueryResult[T]) Empty() bool { return q.bitSet.Count() == 0 }

func (q *QueryResult[T]) Values() []T {
	list := make([]T, 0, q.bitSet.Count())

	q.list.lock.RLock()
	defer q.list.lock.RUnlock()

	q.bitSet.Values(func(r uint32) bool {
		// get from the FreeList without lock
		o, _ := q.list.list.Get(int(r))
		list = append(list, o)

		return true
	})

	return list
}

func (q *QueryResult[T]) Sort(less func(*T, *T) bool) []T {
	list := q.Values()
	sort.Slice(list, func(i, j int) bool { return less(&list[i], &list[j]) })
	return list
}

func (q *QueryResult[T]) Remove() {
	q.list.lock.Lock()
	defer q.list.lock.Unlock()

	q.bitSet.Values(func(r uint32) bool {
		q.list.removeNoLock(int(r))
		return true
	})
}

func (l *IndexList[T]) removeNoLock(index int) (t T, removed bool) {
	item, found := l.list.Get(index)
	if !found {
		return item, found
	}

	removed = l.list.Remove(index)
	l.allIDs.UnSet(uint32(index))

	for _, fieldIndex := range l.fieldIndexMap {
		val := fieldIndex.fieldFn(&item)
		fieldIndex.index.UnSet(val, uint32(index))
	}

	return item, removed
}
