package main

import (
	"fmt"
	"reflect"
)

func NewFieldIndexMap[T any, R Row]() FieldIndexMap[T, R] {
	return make(FieldIndexMap[T, R], 0)
}

type FieldIndexMap[T any, R Row] map[string]struct {
	index             Index[R]
	fieldFn           func(*T) any
	fieldFnResultType reflect.Type
}

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
