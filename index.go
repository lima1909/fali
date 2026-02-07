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

// SortedIndex is well suited for Queries with: Range, Min, Max, Greater and Less
func NewSortedIndex[K cmp.Ordered, R Row]() Index[R] {
	return &SortedIndex[K, R]{skipList: NewSkipList[K, *BitSet[R]]()}
}

type SortedIndex[K cmp.Ordered, R Row] struct{ skipList SkipList[K, *BitSet[R]] }

func (si *SortedIndex[K, R]) Set(value any, row R)   { set(&si.skipList, value, row) }
func (si *SortedIndex[K, R]) UnSet(value any, row R) { unSet(&si.skipList, value, row) }
func (si *SortedIndex[K, R]) Get(relation Relation, value any) *BitSet[R] {
	return get(&si.skipList, relation, value)
}

func set[K cmp.Ordered, R Row](skipList *SkipList[K, *BitSet[R]], value any, row R) {
	bs, found := skipList.Get(value.(K))
	if !found {
		bs = NewBitSet[R]()
	}
	bs.Set(row)
	skipList.Put(value.(K), bs)
}

func unSet[K cmp.Ordered, R Row](skipList *SkipList[K, *BitSet[R]], value any, row R) {
	if bs, found := skipList.Get(value.(K)); found {
		bs.UnSet(row)
		if bs.Count() == 0 {
			skipList.Delete(value.(K))
		}
	}
}

func get[K cmp.Ordered, R Row](skipList *SkipList[K, *BitSet[R]], relation Relation, value any) *BitSet[R] {
	switch relation {
	case Equal:
		bs, found := skipList.Get(value.(K))
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
		skipList.Range(minValue, value.(K), func(key K, bs *BitSet[R]) bool {
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
		skipList.Range(minValue, value.(K), func(key K, bs *BitSet[R]) bool {
			result.Or(bs)
			return true
		})

		return result
	default:
		return NewBitSet[R]()
	}
}
