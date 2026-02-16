package main

import (
	"fmt"
	"sort"
	"sync"
)

// IndexList is a list (slice), which is extended by Indices for fast finding Items in the list.
type IndexList[T any, ID comparable] struct {
	list     FreeList[T]
	indexMap indexMap[T, ID]

	lock sync.RWMutex
}

// NewIndexList create a new IndexList
func NewIndexList[T any]() *IndexList[T, struct{}] {
	return &IndexList[T, struct{}]{
		list:     NewFreeList[T](),
		indexMap: newIndexMap[T, struct{}](nil),
	}
}

// NewIndexList create a new IndexList with an ID-Index
func NewIndexListWithID[T any, ID comparable](fieldIDGetFn func(*T) ID) *IndexList[T, ID] {
	return &IndexList[T, ID]{
		list:     NewFreeList[T](),
		indexMap: newIndexMap(newIDMapIndex(fieldIDGetFn)),
	}
}

// CreateIndex create a new Index:
//   - fieldName: a name for a field of the saved Item
//   - fieldGetFn: a function, which returns the value of an field
//   - Index: a impl of the Index interface
func (l *IndexList[T, ID]) CreateIndex(fieldName string, index Index32[T]) error {
	if fieldName == "" {
		return fmt.Errorf("empty fieldName is not allowed")
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	if _, exist := l.indexMap.index[fieldName]; exist {
		return fmt.Errorf("field-name: %s already exists", fieldName)
	}

	for idx, item := range l.list.Iter() {
		index.Set(&item, uint32(idx))
	}

	l.indexMap.index[fieldName] = index
	return nil
}

// Insert add the given Item to the list,
// There is NO check, for existing this Item in the list, it will ALWAYS inserting!
func (l *IndexList[T, ID]) Insert(item T) int {
	l.lock.Lock()
	defer l.lock.Unlock()

	idx := l.list.Insert(item)
	l.indexMap.Set(&item, idx)

	return idx
}

// Update replaces an item and consistently updates all registered indexes.
func (l *IndexList[T, ID]) Update(item T) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	id, idx, err := l.indexMap.getIDByItem(&item)
	if err != nil {
		return err
	}

	// overwrite the data in the main list
	oldItem, ok := l.list.Set(idx, item)
	if !ok {
		return ErrValueNotFound{id}
	}

	// re-index
	for _, index := range l.indexMap.index {
		// TODO: do it better: check is it neccesary/dirty
		index.UnSet(&oldItem, uint32(idx))
		index.Set(&item, uint32(idx))
	}

	return nil
}

// Remove an item by the given ID.
// This works ONLY, if an ID is defined (with calling: NewIndexListWithID)
// errors:
// - wrong datatype
// - ID not found
// - no ID defined
func (l *IndexList[T, ID]) Remove(id ID) (bool, error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	idx, err := l.indexMap.getIndexByID(id)
	if err != nil {
		return false, err
	}

	_, removed := l.removeNoLock(idx)
	return removed, nil
}

// Get returns an item by the given ID.
// This works ONLY, if an ID is defined (with calling: NewIndexListWithID)
// errors:
// - wrong datatype
// - ID not found
// - no ID defined
func (l *IndexList[T, ID]) Get(id ID) (T, error) {
	l.lock.RLock()
	defer l.lock.RUnlock()

	idx, err := l.indexMap.getIndexByID(id)
	if err != nil {
		var null T
		return null, err
	}

	// not found should be possible
	item, _ := l.list.Get(idx)
	return item, nil
}

// ContainsID check, is this ID found in the list.
func (l *IndexList[T, ID]) ContainsID(id ID) bool {
	l.lock.RLock()
	defer l.lock.RUnlock()

	_, err := l.indexMap.getIndexByID(id)
	return err == nil
}

// Query execute the given Query.
func (l *IndexList[T, ID]) Query(query Query32) (QueryResult[T, ID], error) {
	l.lock.RLock()
	bs, _, err := query(l.indexMap.LookupByName, l.indexMap.allIDs)
	l.lock.RUnlock()

	if err != nil {
		return QueryResult[T, ID]{}, err
	}

	return QueryResult[T, ID]{bitSet: bs, list: l}, nil
}

// Count the Items, which in this list exist
func (l *IndexList[T, ID]) Count() int {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.indexMap.allIDs.Count()
}

//go:inline
func (l *IndexList[T, ID]) removeNoLock(index int) (t T, removed bool) {
	item, found := l.list.Get(index)
	if !found {
		return item, found
	}

	removed = l.list.Remove(index)
	l.indexMap.UnSet(&item, index)

	return item, removed
}

type QueryResult[T any, ID comparable] struct {
	bitSet *BitSet[uint32]
	list   *IndexList[T, ID]
}

func (q *QueryResult[T, ID]) Count() int  { return q.bitSet.Count() }
func (q *QueryResult[T, ID]) Empty() bool { return q.bitSet.Count() == 0 }

func (q *QueryResult[T, ID]) Values() []T {
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

func (q *QueryResult[T, ID]) Sort(less func(*T, *T) bool) []T {
	list := q.Values()
	sort.Slice(list, func(i, j int) bool { return less(&list[i], &list[j]) })
	return list
}

func (q *QueryResult[T, ID]) Remove() {
	q.list.lock.Lock()
	defer q.list.lock.Unlock()

	q.bitSet.Values(func(r uint32) bool {
		q.list.removeNoLock(int(r))
		return true
	})
}
