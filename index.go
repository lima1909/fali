package main

import (
	"cmp"
)

// Index32 the IndexList only supported uint32 List-Indices
type Index32[T any] = Index[T, uint32]

// Index is interface for handling the mapping of an Value: V to an List-Index: LI
// The Value V comes from a func(*OBJ) V
type Index[OBJ any, LI Value] interface {
	Set(*OBJ, LI)
	UnSet(*OBJ, LI)
	Get(Relation, any) (*BitSet[LI], error)
}

// MapIndex is a mapping of any value to the Index in the List.
// This index only supported Queries with the Equal Ralation!
type MapIndex[OBJ any, V any, LI Value] struct {
	data       map[any]*BitSet[LI]
	fieldGetFn func(*OBJ) V
}

func NewMapIndex[OBJ any, V any](fieldGetFn func(*OBJ) V) Index32[OBJ] {
	return &MapIndex[OBJ, V, uint32]{
		data:       make(map[any]*BitSet[uint32]),
		fieldGetFn: fieldGetFn,
	}
}

func (mi *MapIndex[OBJ, V, LI]) Set(obj *OBJ, lidx LI) {
	value := mi.fieldGetFn(obj)
	bs, found := mi.data[value]
	if !found {
		bs = NewBitSet[LI]()
	}
	bs.Set(lidx)
	mi.data[value] = bs
}

func (mi *MapIndex[OBJ, V, LI]) UnSet(obj *OBJ, lidx LI) {
	value := mi.fieldGetFn(obj)
	if bs, found := mi.data[value]; found {
		bs.UnSet(lidx)
		if bs.Count() == 0 {
			delete(mi.data, value)
		}
	}
}

func (mi *MapIndex[OBJ, V, LI]) Get(relation Relation, value any) (*BitSet[LI], error) {
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
type SortedIndex[OBJ any, V cmp.Ordered, LI Value] struct {
	skipList   SkipList[V, *BitSet[LI]]
	fieldGetFn func(*OBJ) V
}

func NewSortedIndex[OBJ any, V cmp.Ordered](fieldGetFn func(*OBJ) V) Index32[OBJ] {
	return &SortedIndex[OBJ, V, uint32]{
		skipList:   NewSkipList[V, *BitSet[uint32]](),
		fieldGetFn: fieldGetFn,
	}
}

func (si *SortedIndex[OBJ, V, LI]) Set(obj *OBJ, lidx LI) {
	value := si.fieldGetFn(obj)
	bs, found := si.skipList.Get(value)
	if !found {
		bs = NewBitSet[LI]()
	}
	bs.Set(lidx)
	si.skipList.Put(value, bs)
}

func (si *SortedIndex[OBJ, V, LI]) UnSet(obj *OBJ, lidx LI) {
	value := si.fieldGetFn(obj)
	if bs, found := si.skipList.Get(value); found {
		bs.UnSet(lidx)
		if bs.Count() == 0 {
			si.skipList.Delete(value)
		}
	}
}

func (si *SortedIndex[OBJ, V, LI]) Get(relation Relation, value any) (*BitSet[LI], error) {
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
			return NewBitSet[LI](), nil
		}

		result := NewBitSet[LI]()
		si.skipList.Range(minValue, value.(V), func(v V, bs *BitSet[LI]) bool {
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
			return NewBitSet[LI](), nil
		}

		result := NewBitSet[LI]()
		si.skipList.Range(minValue, value.(V), func(_ V, bs *BitSet[LI]) bool {
			result.Or(bs)
			return true
		})

		return result, nil
	default:
		return nil, ErrInvalidRelation{relation}
	}
}
