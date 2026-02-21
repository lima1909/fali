package main

import (
	"cmp"
	"fmt"
	"reflect"
	"unsafe"
)

// fieldIndexMap maps a given field name to an Index
type indexMap[OBJ any, ID comparable] struct {
	idIndex idIndex[OBJ, ID]
	index   map[string]Index32[OBJ]
	allIDs  *BitSet[uint32]
}

func newIndexMap[OBJ any, ID comparable](idIndex idIndex[OBJ, ID]) indexMap[OBJ, ID] {
	return indexMap[OBJ, ID]{
		idIndex: idIndex,
		index:   make(map[string]Index32[OBJ]),
		allIDs:  NewBitSet[uint32](),
	}
}

// LookupByName finds the Lookup by a given field-name
func (i indexMap[OBJ, ID]) LookupByName(fieldName string) (Lookup32, error) {
	if fieldName == "" {
		if i.idIndex == nil {
			return nil, ErrNoIdIndexDefined{}
		}
		return i.idIndex, nil
	}

	if idx, found := i.index[fieldName]; found {
		return idx, nil
	}

	return nil, ErrInvalidIndexdName{fieldName}
}

func (i indexMap[OBJ, ID]) Set(obj *OBJ, idx int) {
	if i.idIndex != nil {
		i.idIndex.Set(obj, idx)
	}

	uidx := uint32(idx)
	i.allIDs.Set(uidx)
	for _, fieldIndex := range i.index {
		fieldIndex.Set(obj, uidx)
	}
}

func (i indexMap[OBJ, ID]) UnSet(obj *OBJ, idx int) {
	if i.idIndex != nil {
		i.idIndex.UnSet(obj, idx)
	}

	uidx := uint32(idx)
	i.allIDs.UnSet(uidx)
	for _, fieldIndex := range i.index {
		fieldIndex.UnSet(obj, uidx)
	}
}

func (i indexMap[OBJ, ID]) getIndexByID(id ID) (int, error) {
	if i.idIndex == nil {
		return 0, ErrNoIdIndexDefined{}
	}

	return i.idIndex.GetIndex(id)
}

func (i indexMap[OBJ, ID]) getIDByItem(item *OBJ) (ID, int, error) {
	if i.idIndex == nil {
		var id ID
		return id, 0, ErrNoIdIndexDefined{}
	}

	return i.idIndex.GetID(item)
}

type idIndex[OBJ any, ID comparable] interface {
	Set(*OBJ, int)
	UnSet(*OBJ, int)
	GetIndex(ID) (int, error)
	GetID(*OBJ) (ID, int, error)
	Lookup32
}

type idMapIndex[OBJ any, ID comparable] struct {
	data       map[ID]int
	fieldGetFn FromField[OBJ, ID]
}

func newIDMapIndex[OBJ any, ID comparable](fieldGetFn FromField[OBJ, ID]) idIndex[OBJ, ID] {
	return &idMapIndex[OBJ, ID]{
		data:       make(map[ID]int),
		fieldGetFn: fieldGetFn,
	}
}

func (mi *idMapIndex[OBJ, ID]) Set(obj *OBJ, lidx int) {
	id := mi.fieldGetFn(obj)
	mi.data[id] = lidx
}

func (mi *idMapIndex[OBJ, ID]) UnSet(obj *OBJ, lidx int) {
	id := mi.fieldGetFn(obj)
	delete(mi.data, id)
}

func (mi *idMapIndex[OBJ, ID]) GetIndex(id ID) (int, error) {
	if lidx, found := mi.data[id]; found {
		return lidx, nil
	}

	return 0, ErrValueNotFound{id}
}

func (mi *idMapIndex[OBJ, ID]) GetID(item *OBJ) (ID, int, error) {
	id := mi.fieldGetFn(item)
	if lidx, found := mi.data[id]; found {
		return id, lidx, nil
	}

	var null ID
	return null, 0, ErrValueNotFound{id}
}

func (mi *idMapIndex[OBJ, ID]) Get(relation Relation, value any) (*BitSet[uint32], error) {
	if _, ok := value.(ID); !ok {
		return nil, ErrInvalidIndexValue[ID]{value}
	}

	if relation != Equal {
		return nil, ErrInvalidRelation{relation}
	}

	idx, err := mi.GetIndex(value.(ID))
	if err != nil {
		return nil, err
	}

	return NewBitSetFrom(uint32(idx)), nil

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

// FromField is a function, which returns a value from an given object.
// example:
// Person{name string}
// func (p *Person) Name() { return p.name }
// (*Person).Name is the FieldGetFn
type FromField[OBJ any, V any] = func(*OBJ) V

// FromValue returns a Getter that simply returns the value itself.
// Use this when your list contains the raw values you want to index.
func FromValue[V any]() FromField[V, V] { return func(v *V) V { return *v } }

// FromName returns per reflection the propery (field) value from the given object.
func FromName[OBJ any, V any](fieldName string) FromField[OBJ, V] {
	var zero OBJ
	typ := reflect.TypeOf(zero)
	isPtr := false
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
		isPtr = true
	}

	if typ.Kind() != reflect.Struct {
		panic(fmt.Sprintf("expected struct, got %s", typ.Kind()))
	}

	field, ok := typ.FieldByName(fieldName)
	if !ok {
		panic(fmt.Sprintf("field %s not found", fieldName))
	}
	// reflection cannot access lowercase (unexported) fields via .Interface()
	// unless we use unsafe, but let's stick to standard safety checks at setup time.
	// Actually, unsafe access works on unexported fields too, but usually discouraged.
	// But let's fail as per original behavior.
	if !field.IsExported() {
		panic(fmt.Sprintf("field %s is unexported", fieldName))
	}

	offset := field.Offset

	if isPtr {
		// OBJ is *Struct. input is **Struct.
		return func(obj *OBJ) V {
			// *obj is the *Struct.
			// We need unsafe.Pointer(*obj) + offset
			structPtr := *(**unsafe.Pointer)(unsafe.Pointer(obj))
			if structPtr == nil {
				var zero V
				return zero // Or panic? Original reflect would panic on nil pointer deref usually.
			}
			return *(*V)(unsafe.Add(*structPtr, offset))
		}
	}

	// OBJ is Struct. input is *Struct.
	return func(obj *OBJ) V {
		// obj is *Struct
		return *(*V)(unsafe.Add(unsafe.Pointer(obj), offset))
	}
}

// MapIndex is a mapping of any value to the Index in the List.
// This index only supported Queries with the Equal Ralation!
type MapIndex[OBJ any, V any, LI Value] struct {
	data       map[any]*BitSet[LI]
	fieldGetFn FromField[OBJ, V]
}

func NewMapIndex[OBJ any, V any](fromField FromField[OBJ, V]) Index32[OBJ] {
	return &MapIndex[OBJ, V, uint32]{
		data:       make(map[any]*BitSet[uint32]),
		fieldGetFn: fromField,
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
		return NewBitSet[LI](), nil
	}

	return bs, nil
}

// SortedIndex is well suited for Queries with: Range, Min, Max, Greater and Less
type SortedIndex[OBJ any, V cmp.Ordered, LI Value] struct {
	skipList   SkipList[V, *BitSet[LI]]
	fieldGetFn FromField[OBJ, V]
}

func NewSortedIndex[OBJ any, V cmp.Ordered](fieldGetFn FromField[OBJ, V]) Index32[OBJ] {
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
		return NewBitSet[LI](), nil
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
	case StartsWith:
		if _, ok := value.(string); !ok {
			return nil, ErrInvalidIndexValue[string]{value}
		}

		result := NewBitSet[LI]()
		si.skipList.StringStartsWith(value.(V), func(_ V, bs *BitSet[LI]) bool {
			result.Or(bs)
			return true
		})
		return result, nil
	default:
		return nil, ErrInvalidRelation{relation}
	}
}
