package main

import (
	"cmp"
)

func NewFieldIndexMap[T any, R Row]() FieldIndexMap[T, R] {
	return make(FieldIndexMap[T, R], 0)
}

// FieldIndexMap maps a given field name to an Index
type FieldIndexMap[T any, R Row] map[string]Index[T, R]

// IndexByName is the default impl for the FieldIndexFn
func (f FieldIndexMap[T, R]) IndexByName(fieldName string, val any) (QueryFieldGetFn[R], error) {
	if idx, found := f[fieldName]; found {
		return idx.Get, nil
	}
	return nil, ErrInvalidIndexdName{fieldName}
}

type Row = Value

type Index[T any, R Row] interface {
	Set(T, R)
	UnSet(T, R)
	Get(Relation, any) (*BitSet[R], error)
}

type FieldGetFn[T any, V any] func(T) V

// MapIndex is a mapping of any value to the Index in the List.
// This index only supported Queries with the Equal Ralation!
type MapIndex[T any, V any, R Row] struct {
	data       map[any]*BitSet[R]
	fieldGetFn FieldGetFn[T, V]
}

func NewMapIndex[T any, V any](fieldGetFn FieldGetFn[T, V]) *MapIndex[T, V, uint32] {
	return &MapIndex[T, V, uint32]{
		data:       make(map[any]*BitSet[uint32]),
		fieldGetFn: fieldGetFn,
	}
}

func (mi *MapIndex[T, V, R]) Set(obj T, row R) {
	value := mi.fieldGetFn(obj)
	bs, found := mi.data[value]
	if !found {
		bs = NewBitSet[R]()
	}
	bs.Set(row)
	mi.data[value] = bs
}

func (mi *MapIndex[T, V, R]) UnSet(obj T, row R) {
	value := mi.fieldGetFn(obj)
	if bs, found := mi.data[value]; found {
		bs.UnSet(row)
		if bs.Count() == 0 {
			delete(mi.data, value)
		}
	}
}

func (mi *MapIndex[T, V, R]) Get(relation Relation, value any) (*BitSet[R], error) {
	if _, ok := value.(V); !ok {
		return nil, ErrInvalidIndexValue[V]{value}
	}

	if relation != Equal {
		return nil, ErrInvalidRelation{relation}
	}

	bs, found := mi.data[value]
	if !found {
		return nil, ErrValueNotFound{value}
	}

	return bs, nil
}

// SortedIndex is well suited for Queries with: Range, Min, Max, Greater and Less
type SortedIndex[T any, V cmp.Ordered, R Row] struct {
	skipList   SkipList[V, *BitSet[R]]
	fieldGetFn FieldGetFn[T, V]
}

func NewSortedIndex[T any, V cmp.Ordered](fieldGetFn FieldGetFn[T, V]) Index[T, uint32] {
	return &SortedIndex[T, V, uint32]{
		skipList:   NewSkipList[V, *BitSet[uint32]](),
		fieldGetFn: fieldGetFn,
	}
}

func (si *SortedIndex[T, V, R]) Set(obj T, row R) {
	value := si.fieldGetFn(obj)
	bs, found := si.skipList.Get(value)
	if !found {
		bs = NewBitSet[R]()
	}
	bs.Set(row)
	si.skipList.Put(value, bs)
}

func (si *SortedIndex[T, V, R]) UnSet(obj T, row R) {
	value := si.fieldGetFn(obj)
	if bs, found := si.skipList.Get(value); found {
		bs.UnSet(row)
		if bs.Count() == 0 {
			si.skipList.Delete(value)
		}
	}
}

func (si *SortedIndex[T, V, R]) Get(relation Relation, value any) (*BitSet[R], error) {
	if _, ok := value.(V); !ok {
		return nil, ErrInvalidIndexValue[V]{value}
	}

	switch relation {
	case Equal:
		bs, found := si.skipList.Get(value.(V))
		if !found {
			return nil, ErrValueNotFound{value}
		}
		return bs, nil
	case Less:
		minValue, found := si.skipList.MinKey()
		if !found {
			return NewBitSet[R](), nil
		}

		result := NewBitSet[R]()
		si.skipList.Range(minValue, value.(V), func(v V, bs *BitSet[R]) bool {
			if v == value {
				return false
			}

			result.Or(bs)
			return true
		})

		return result, nil
	case LessEqual:
		minValue, found := si.skipList.MinKey()
		if !found {
			return NewBitSet[R](), nil
		}

		result := NewBitSet[R]()
		si.skipList.Range(minValue, value.(V), func(_ V, bs *BitSet[R]) bool {
			result.Or(bs)
			return true
		})

		return result, nil
	default:
		return nil, ErrInvalidRelation{relation}
	}
}
