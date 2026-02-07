package main

import (
	"cmp"
	"fmt"
	"reflect"
)

func NewFieldIndexMap[T any, R Row]() FieldIndexMap[T, R] {
	return make(FieldIndexMap[T, R], 0)
}

type FieldGetFn[T any] = func(*T) any

type FieldIndex[T any, R Row] struct {
	index             Index[R]
	fieldFn           FieldGetFn[T]
	fieldFnResultType reflect.Type
}

type FieldIndexMap[T any, R Row] map[string]FieldIndex[T, R]

// IndexByName is the default impl for the FieldIndexFn
func (f FieldIndexMap[T, R]) IndexByName(fieldName string, val any) (Index[R], error) {
	if idx, found := f[fieldName]; found {
		if idx.fieldFnResultType != reflect.TypeOf(val) {
			return nil, fmt.Errorf("invalid index value type: %s, expected type: %s", val, idx.fieldFnResultType)
		}

		return idx.index, nil
	}

	return nil, fmt.Errorf("could not found index for field name: %s", fieldName)
}

type Relation int8

const (
	Equal Relation = 1 << iota
	Less
	Greater
	LessEqual
	GreaterEqual
)

type Row = Value

type Index[R Row] interface {
	Set(any, R)
	UnSet(any, R)
	Get(Relation, any) *BitSet[R]
}

// MapIndex is a mapping of any value to the Index in the List.
// This index only supported Queries with the Equal Ralation!
type MapIndex[R Row] struct {
	data map[any]*BitSet[R]
}

func NewMapIndex[R Row]() *MapIndex[R] {
	return &MapIndex[R]{data: make(map[any]*BitSet[R])}
}

func (mi *MapIndex[R]) Set(value any, row R) {
	bs, found := mi.data[value]
	if !found {
		bs = NewBitSet[R]()
	}
	bs.Set(row)
	mi.data[value] = bs
}

func (mi *MapIndex[R]) UnSet(value any, row R) {
	if bs, found := mi.data[value]; found {
		bs.UnSet(row)
		if bs.Count() == 0 {
			delete(mi.data, value)
		}
	}
}

func (mi *MapIndex[R]) Get(relation Relation, value any) *BitSet[R] {
	if relation != Equal {
		//TODO: better return an error
		return NewBitSet[R]()
	}

	bs, found := mi.data[value]
	if !found {
		return NewBitSet[R]()
	}

	return bs
}

// ------------------------
type Uint8SortedIndex[R Row] SortedIndex[uint8, R]

func NewUint8SortedIndex[R Row]() *Uint8SortedIndex[R] {
	return &Uint8SortedIndex[R]{skipList: NewSkipList[uint8, *BitSet[R]]()}
}
func (si *Uint8SortedIndex[R]) Set(value any, row R)   { Set(&si.skipList, value.(uint8), row) }
func (si *Uint8SortedIndex[R]) UnSet(value any, row R) { UnSet(&si.skipList, value.(uint8), row) }
func (si *Uint8SortedIndex[R]) Get(relation Relation, value any) *BitSet[R] {
	return Get(&si.skipList, relation, value.(uint8))
}

// ------------------------
// SortedIndex is well suited for Queries with: Range, Min, Max, Greater and Less
type SortedIndex[K cmp.Ordered, R Row] struct{ skipList SkipList[K, *BitSet[R]] }

func NewSortedIndex[R Row]() *SortedIndex[uint, R] {
	return &SortedIndex[uint, R]{skipList: NewSkipList[uint, *BitSet[R]]()}
}

func (si SortedIndex[K, R]) Set(value K, row R)   { Set(&si.skipList, value, row) }
func (si SortedIndex[K, R]) UnSet(value K, row R) { UnSet(&si.skipList, value, row) }
func (si SortedIndex[K, R]) Get(relation Relation, value K) *BitSet[R] {
	return Get(&si.skipList, relation, value)
}

func Set[K cmp.Ordered, R Row](skipList *SkipList[K, *BitSet[R]], value K, row R) {
	bs, found := skipList.Get(value)
	if !found {
		bs = NewBitSet[R]()
	}
	bs.Set(row)
	skipList.Put(value, bs)
}

func UnSet[K cmp.Ordered, R Row](skipList *SkipList[K, *BitSet[R]], value K, row R) {
	if bs, found := skipList.Get(value); found {
		bs.UnSet(row)
		if bs.Count() == 0 {
			skipList.Delete(value)
		}
	}
}

func Get[K cmp.Ordered, R Row](skipList *SkipList[K, *BitSet[R]], relation Relation, value K) *BitSet[R] {
	switch relation {
	case Equal:
		bs, found := skipList.Get(value)
		if !found {
			return NewBitSet[R]()
		}
		return bs
	case Less:
		minValue, found := skipList.MinKey()
		if !found {
			return NewBitSet[R]()
		}

		result := NewBitSet[R]()
		skipList.Range(minValue, value, func(key K, bs *BitSet[R]) bool {
			if key == value {
				return false
			}

			result.Or(bs)
			return true
		})

		return result
	case LessEqual:
		minValue, found := skipList.MinKey()
		if !found {
			return NewBitSet[R]()
		}

		result := NewBitSet[R]()
		skipList.Range(minValue, value, func(key K, bs *BitSet[R]) bool {
			result.Or(bs)
			return true
		})

		return result
	default:
		return NewBitSet[R]()
	}
}
