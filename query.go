package main

type Relation int8

const (
	Equal Relation = 1 << iota
	Less
	Greater
	LessEqual
	GreaterEqual
	// NotEqual       Relation = "!="
)

type IndexKey struct {
	inum    int
	fnum    float64
	text    string
	boolean bool
}

func Int(i int) IndexKey       { return IndexKey{inum: i} }
func Float(f float64) IndexKey { return IndexKey{fnum: f} }
func Bool(b bool) IndexKey     { return IndexKey{boolean: b} }
func String(s string) IndexKey { return IndexKey{text: s} }

type Row = Value

type Index[R Row] interface {
	Set(IndexKey, R)
	UnSet(IndexKey, R)
	Get(Relation, IndexKey) *BitSet[R]
}

type MapIndex[R Row] struct {
	data map[IndexKey]*BitSet[R]
}

func NewMapIndex[R Row]() *MapIndex[R] {
	return &MapIndex[R]{data: make(map[IndexKey]*BitSet[R])}
}

func (idx *MapIndex[R]) Set(key IndexKey, row R) {
	bs, found := idx.data[key]
	if !found {
		bs = NewBitSet[R]()
	}
	bs.Set(row)
	idx.data[key] = bs
}

func (idx *MapIndex[R]) UnSet(key IndexKey, row R) {
	if bs, found := idx.data[key]; found {
		bs.UnSet(row)
		if bs.Count() == 0 {
			delete(idx.data, key)
		}
	}
}

func (idx *MapIndex[R]) Get(relation Relation, key IndexKey) *BitSet[R] {
	if relation != Equal {
		return NewBitSet[R]()
	}

	bs, found := idx.data[key]
	if !found {
		return NewBitSet[R]()
	}

	return bs
}

// FieldIndex is a mapping from a given (field) name to an Index
type FieldIndex[R Row] map[string]Index[R]

// Query is a filter function, find the correct Index an execute the Index.Get method
// and returns a BitSet pointer
type Query[R Row] func(fi FieldIndex[R], allIDs *BitSet[R]) (_ *BitSet[R], canMutate bool)

func Eq[R Row](key string, val IndexKey) Query[R] {
	return func(fi FieldIndex[R], _ *BitSet[R]) (*BitSet[R], bool) {
		idx, ok := fi[key]
		if !ok {
			return NewBitSet[R](), true
		}

		return idx.Get(Equal, val), false
	}
}

func Not[R Row](q Query[R]) Query[R] {
	return func(fi FieldIndex[R], allIDs *BitSet[R]) (_ *BitSet[R], canMutate bool) {
		qres := ensureMutable(q(fi, allIDs))

		// maybe i can change the copy?
		result := allIDs.Copy()
		result.AndNot(qres)
		return result, true
	}
}

func (q Query[R]) And(other Query[R]) Query[R] {
	return func(fi FieldIndex[R], allIDs *BitSet[R]) (_ *BitSet[R], canMutate bool) {
		result := ensureMutable(q(fi, allIDs))
		right, _ := other(fi, allIDs)

		result.And(right)
		return result, true
	}
}

func (q Query[R]) Or(other Query[R]) Query[R] {
	return func(fi FieldIndex[R], allIDs *BitSet[R]) (_ *BitSet[R], canMutate bool) {
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
