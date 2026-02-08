package main

type Relation int8

const (
	Equal Relation = 1 << iota
	Less
	Greater
	LessEqual
	GreaterEqual
)

type QueryFieldGetFn[R Row] = func(Relation, any) (*BitSet[R], error)

// FieldIndexFn is the interfact to the FieldIndexMap
type FieldIndexFn[R Row] = func(string, any) (QueryFieldGetFn[R], error)

// Query is a filter function, find the correct Index an execute the Index.Get method
// and returns a BitSet pointer
type Query[R Row] func(fi FieldIndexFn[R], allIDs *BitSet[R]) (bs *BitSet[R], canMutate bool, err error)

// Query32 supported only  uint32 List-Indices
type Query32 = Query[uint32]

// All means returns all Items, no filtering
func All() Query32 { return all[uint32]() }

//go:inline
func all[R Row]() Query[R] {
	return func(_ FieldIndexFn[R], allIDs *BitSet[R]) (_ *BitSet[R], canMutate bool, _ error) {
		return allIDs, false, nil
	}
}

// Rel fieldName rel (Equal, Less, ...) val
func Rel(fieldName string, r Relation, val any) Query32 { return rel[uint32](fieldName, r, val) }

//go:inline
func rel[R Row](fieldName string, relation Relation, val any) Query[R] {
	return func(fi FieldIndexFn[R], _ *BitSet[R]) (_ *BitSet[R], canMutate bool, _ error) {
		get, err := fi(fieldName, val)
		if err != nil {
			return nil, false, err
		}

		bs, err := get(relation, val)
		return bs, false, err
	}
}

// Eq fieldName = val
func Eq(fieldName string, val any) Query32 { return eq[uint32](fieldName, val) }

//go:inline
func eq[R Row](fieldName string, val any) Query[R] {
	return func(fi FieldIndexFn[R], _ *BitSet[R]) (_ *BitSet[R], canMutate bool, _ error) {
		get, err := fi(fieldName, val)
		if err != nil {
			return nil, false, err
		}

		bs, err := get(Equal, val)
		return bs, false, err
	}
}

// In combines Eq with an Or
// In("name", "Paul", "Egon") => name == "Paul" Or name == "Egon"
func In(fieldName string, vals ...any) Query32 { return in[uint32](fieldName, vals...) }

//go:inline
func in[R Row](fieldName string, vals ...any) Query[R] {
	return func(fi FieldIndexFn[R], _ *BitSet[R]) (_ *BitSet[R], canMutate bool, _ error) {
		if len(vals) == 0 {
			return NewBitSet[R](), true, nil
		}

		get, err := fi(fieldName, vals[0])
		if err != nil {
			return nil, false, err
		}

		bs, err := get(Equal, vals[0])
		if err != nil {
			return nil, false, err
		}

		if len(vals) == 1 {
			return bs, false, nil
		}

		bs = bs.Copy()
		for _, val := range vals[1:] {
			bsGet, err := get(Equal, val)
			if err != nil {
				return nil, false, err
			}
			bs.Or(bsGet)
		}

		return bs, true, nil
	}
}

// NotEq is a shorcut for Not(Eq(...))
func NotEq(fieldName string, val any) Query32 { return notEq[uint32](fieldName, val) }

//go:inline
func notEq[R Row](fieldName string, val any) Query[R] {
	return func(fi FieldIndexFn[R], allIDs *BitSet[R]) (_ *BitSet[R], canMutate bool, _ error) {
		eq := eq[R](fieldName, val)
		return Not(eq)(fi, allIDs)
	}
}

// Not Not(Query)
func Not[R Row](q Query[R]) Query[R] {
	return func(fi FieldIndexFn[R], allIDs *BitSet[R]) (_ *BitSet[R], canMutate bool, _ error) {
		// can Mutate is not relevant, because allIDs are copied
		qres, _, err := q(fi, allIDs)
		if err != nil {
			return nil, false, err
		}

		// maybe i can change the copy?
		result := allIDs.Copy()
		result.AndNot(qres)
		return result, true, nil
	}
}

// And Query and Query
func (q Query[R]) And(other Query[R]) Query[R] {
	return func(fi FieldIndexFn[R], allIDs *BitSet[R]) (_ *BitSet[R], canMutate bool, _ error) {
		result, err := ensureMutable(q(fi, allIDs))
		if err != nil {
			return nil, false, err
		}
		right, _, err := other(fi, allIDs)
		if err != nil {
			return nil, false, err
		}

		result.And(right)
		return result, true, nil
	}
}

// Or Query or Query
func (q Query[R]) Or(other Query[R]) Query[R] {
	return func(fi FieldIndexFn[R], allIDs *BitSet[R]) (_ *BitSet[R], canMutate bool, _ error) {
		result, err := ensureMutable(q(fi, allIDs))
		if err != nil {
			return nil, false, err
		}
		right, _, err := other(fi, allIDs)
		if err != nil {
			return nil, false, err
		}

		result.Or(right)
		return result, true, nil
	}
}

// check, must the BitSet copied or not
// only copy, if not mutable
//
//go:inline
func ensureMutable[R Row](b *BitSet[R], canMutate bool, err error) (*BitSet[R], error) {
	if err != nil {
		return nil, err
	}

	if canMutate {
		return b, nil
	}

	return b.Copy(), nil
}
