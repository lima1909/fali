package main

import (
	"unsafe"
)

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
		return NewBitSet[R]()
	}

	bs, found := mi.data[value]
	if !found {
		return NewBitSet[R]()
	}

	return bs
}

type FieldIndex[T any, R Row] struct {
	index      Index[R]
	fieldFn    func(*T) any
	resultType uintptr
}

type FieldIndexMap[T any, R Row] map[string]FieldIndex[T, R]

func NewFieldIndexMap[T any, R Row]() FieldIndexMap[T, R] {
	return make(FieldIndexMap[T, R], 0)
}

type FieldIndexFn[R Row] = func(string, any) (Index[R], bool)

// IndexByName is the default impl for the FieldIndexFn
func (f FieldIndexMap[T, R]) IndexByName(fieldName string, val any) (Index[R], bool) {
	if idx, found := f[fieldName]; found {
		//TODO: handle different types, error?
		//
		// typVal := typeID(&val)
		// fmt.Println("---===", typVal == idx.resultType, val)

		return idx.index, true
	}

	return nil, false
}

// Query is a filter function, find the correct Index an execute the Index.Get method
// and returns a BitSet pointer
type Query[R Row] func(fi FieldIndexFn[R], allIDs *BitSet[R]) (_ *BitSet[R], canMutate bool)

func Eq[R Row](fieldName string, val any) Query[R] {
	return func(fi FieldIndexFn[R], _ *BitSet[R]) (_ *BitSet[R], canMutate bool) {
		idx, ok := fi(fieldName, val)
		if !ok {
			return NewBitSet[R](), true
		}

		return idx.Get(Equal, val), false
	}
}

// NotEq is a shorcut for Not(Eq(...))
func NotEq[R Row](fieldName string, val any) Query[R] {
	return func(fi FieldIndexFn[R], allIDs *BitSet[R]) (_ *BitSet[R], canMutate bool) {
		eq := Eq[R](fieldName, val)
		return Not(eq)(fi, allIDs)
	}
}

// In combines Eq with an Or
// In("name", "Paul", "Egon") => name == "Paul" Or name == "Egon"
func In[R Row](fieldName string, vals ...any) Query[R] {
	return func(fi FieldIndexFn[R], _ *BitSet[R]) (_ *BitSet[R], canMutate bool) {
		if len(vals) == 0 {
			return NewBitSet[R](), true
		}

		idx, ok := fi(fieldName, vals[0])
		if !ok {
			return NewBitSet[R](), true
		}

		bs := idx.Get(Equal, vals[0])
		if len(vals) == 1 {
			return bs, false
		}

		bs = bs.Copy()
		for _, val := range vals[1:] {
			bs.Or(idx.Get(Equal, val))
		}

		return bs, true
	}
}

func Not[R Row](q Query[R]) Query[R] {
	return func(fi FieldIndexFn[R], allIDs *BitSet[R]) (_ *BitSet[R], canMutate bool) {
		// can Mutate is not relevant, because allIDs are copied
		qres, _ := q(fi, allIDs)

		// maybe i can change the copy?
		result := allIDs.Copy()
		result.AndNot(qres)
		return result, true
	}
}

func (q Query[R]) And(other Query[R]) Query[R] {
	return func(fi FieldIndexFn[R], allIDs *BitSet[R]) (_ *BitSet[R], canMutate bool) {
		result := ensureMutable(q(fi, allIDs))
		right, _ := other(fi, allIDs)

		result.And(right)
		return result, true
	}
}

func (q Query[R]) Or(other Query[R]) Query[R] {
	return func(fi FieldIndexFn[R], allIDs *BitSet[R]) (_ *BitSet[R], canMutate bool) {
		result := ensureMutable(q(fi, allIDs))
		right, _ := other(fi, allIDs)

		result.Or(right)
		return result, true
	}
}

// check, must the BitSet copied or not
// only copy, if not mutable
//
//go:inline
func ensureMutable[R Row](b *BitSet[R], canMutate bool) *BitSet[R] {
	if canMutate {
		return b
	}

	return b.Copy()
}

// TypeID returns the internal address of the type descriptor
//
//go:inline
func typeID(v any) uintptr {
	return *(*uintptr)(unsafe.Pointer(&v))
}
