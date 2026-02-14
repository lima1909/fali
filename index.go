package main

import (
	"cmp"
)

// fieldIndexMap maps a given field name to an Index
type indexMap[OBJ any] struct {
	idIndex idIndex[OBJ]
	index   map[string]Index32[OBJ]
	allIDs  *BitSet[uint32]
}

func newIndexMap[OBJ any](idIndex idIndex[OBJ]) indexMap[OBJ] {
	return indexMap[OBJ]{
		idIndex: idIndex,
		index:   make(map[string]Index32[OBJ]),
		allIDs:  NewBitSet[uint32](),
	}
}

// LookupByName finds the Lookup by a given field-name
func (i indexMap[OBJ]) LookupByName(fieldName string) (Lookup32, error) {
	if idx, found := i.index[fieldName]; found {
		return idx, nil
	}
	return nil, ErrInvalidIndexdName{fieldName}
}

func (i indexMap[OBJ]) Set(obj *OBJ, idx int) {
	if i.idIndex != nil {
		i.idIndex.Set(obj, idx)
	}

	uidx := uint32(idx)
	i.allIDs.Set(uidx)
	for _, fieldIndex := range i.index {
		fieldIndex.Set(obj, uidx)
	}
}

func (i indexMap[OBJ]) UnSet(obj *OBJ, idx int) {
	if i.idIndex != nil {
		i.idIndex.UnSet(obj, idx)
	}

	uidx := uint32(idx)
	i.allIDs.UnSet(uidx)
	for _, fieldIndex := range i.index {
		fieldIndex.UnSet(obj, uidx)
	}
}

func (i indexMap[OBJ]) getIndexByID(value any) (int, error) {
	if i.idIndex == nil {
		return 0, ErrNoIdIndexDefined{}
	}

	return i.idIndex.Get(value)
}

type idIndex[OBJ any] interface {
	Set(*OBJ, int)
	UnSet(*OBJ, int)
	Get(any) (int, error)
}

type idMapIndex[OBJ any, V any] struct {
	data       map[any]int
	fieldGetFn func(*OBJ) V
}

func newIDMapIndex[OBJ any, V any](fieldGetFn func(*OBJ) V) idIndex[OBJ] {
	return &idMapIndex[OBJ, V]{
		data:       make(map[any]int),
		fieldGetFn: fieldGetFn,
	}
}

func (mi *idMapIndex[OBJ, V]) Set(obj *OBJ, lidx int) {
	value := mi.fieldGetFn(obj)
	mi.data[value] = lidx
}

func (mi *idMapIndex[OBJ, V]) UnSet(obj *OBJ, lidx int) {
	value := mi.fieldGetFn(obj)
	delete(mi.data, value)
}

func (mi *idMapIndex[OBJ, V]) Get(value any) (int, error) {
	if _, ok := value.(V); !ok {
		return 0, ErrInvalidIndexValue[V]{value}
	}

	if lidx, found := mi.data[value]; found {
		return lidx, nil
	}

	return 0, ErrValueNotFound{value}
}

// ------------------------------------------
// here starts the Index with the Index impls
// ------------------------------------------

// Index32 the IndexList only supports uint32 List-Indices
type Index32[T any] = Index[T, uint32]

// Index is interface for handling the mapping of an Value: V to an List-Index: LI
// The Value V comes from a func(*OBJ) V
type Index[OBJ any, LI Value] interface {
	Set(*OBJ, LI)
	UnSet(*OBJ, LI)
	Lookup[LI]
}

// Lookup32 the IndexList only supports uint32 List-Indices
type Lookup32 = Lookup[uint32]

// Lookup returns the BitSet or an error by a given Relation and Value
type Lookup[LI Value] interface {
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
		if bs, found := si.skipList.Get(value.(V)); found {
			return bs, nil
		}
		return nil, ErrValueNotFound{value}
	case Less:
		result := NewBitSet[LI]()
		si.skipList.Less(value.(V), func(v V, bs *BitSet[LI]) bool {
			result.Or(bs)
			return true
		})
		return result, nil
	case LessEqual:
		result := NewBitSet[LI]()
		si.skipList.LessEqual(value.(V), func(_ V, bs *BitSet[LI]) bool {
			result.Or(bs)
			return true
		})
		return result, nil
	case Greater:
		result := NewBitSet[LI]()
		si.skipList.Greater(value.(V), func(v V, bs *BitSet[LI]) bool {
			result.Or(bs)
			return true
		})
		return result, nil
	case GreaterEqual:
		result := NewBitSet[LI]()
		si.skipList.GreaterEqual(value.(V), func(_ V, bs *BitSet[LI]) bool {
			result.Or(bs)
			return true
		})
		return result, nil
	default:
		return nil, ErrInvalidRelation{relation}
	}
}
