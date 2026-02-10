package main

type Relation int8

const (
	Equal = 1 << iota
	Less
	LessEqual
	Greater
	GreaterEqual
)

// Query32 supported only  uint32 List-Indices
type Query32 = Query[uint32]

type QueryFieldGetFn[LI Value] = func(Relation, any) (*BitSet[LI], error)

// FieldIndexFn is the interfact to the FieldIndexMap
type FieldIndexFn[LI Value] = func(string, any) (QueryFieldGetFn[LI], error)

// Query is a filter function, find the correct Index an execute the Index.Get method
// and returns a BitSet pointer
type Query[LI Value] func(fi FieldIndexFn[LI], allIDs *BitSet[LI]) (bs *BitSet[LI], canMutate bool, err error)

// All means returns all Items, no filtering
func All() Query32 { return all[uint32]() }

//go:inline
func all[LI Value]() Query[LI] {
	return func(_ FieldIndexFn[LI], allIDs *BitSet[LI]) (_ *BitSet[LI], canMutate bool, _ error) {
		return allIDs, false, nil
	}
}

// Rel fieldName rel (Equal, Less, ...) val
func Rel(fieldName string, r Relation, val any) Query32 { return rel[uint32](fieldName, r, val) }

//go:inline
func rel[LI Value](fieldName string, relation Relation, val any) Query[LI] {
	return func(fi FieldIndexFn[LI], _ *BitSet[LI]) (_ *BitSet[LI], canMutate bool, _ error) {
		get, err := fi(fieldName, val)
		if err != nil {
			return nil, false, err
		}

		bs, err := get(relation, val)
		return bs, false, err
	}
}

// Eq fieldName = val
func Eq(fieldName string, val any) Query32 {
	return rel[uint32](fieldName, Equal, val)
}

// Lt Less fieldName < val
func Lt(fieldName string, val any) Query32 {
	return rel[uint32](fieldName, Less, val)
}

// Le Less Equal fieldName <= val
func Le(fieldName string, val any) Query32 {
	return rel[uint32](fieldName, LessEqual, val)
}

// Gt Greater fieldName > val
func Gt(fieldName string, val any) Query32 {
	return rel[uint32](fieldName, Greater, val)
}

// Ge Greater Equal fieldName >= val
func Ge(fieldName string, val any) Query32 {
	return rel[uint32](fieldName, GreaterEqual, val)
}

// IsNil is a Query which checks for a given type the nil value
func IsNil[V any](fieldName string) Query32 {
	return isNil[V, uint32](fieldName)
}

//go:inline
func isNil[V any, LI Value](fieldName string) Query[LI] {
	return func(fi FieldIndexFn[LI], _ *BitSet[LI]) (_ *BitSet[LI], canMutate bool, _ error) {
		get, err := fi(fieldName, (*V)(nil))
		if err != nil {
			return nil, false, err
		}

		bs, err := get(Equal, (*V)(nil))
		return bs, false, err
	}
}

// In combines Eq with an Or
// In("name", "Paul", "Egon") => name == "Paul" Or name == "Egon"
func In(fieldName string, vals ...any) Query32 {
	return in[uint32](fieldName, vals...)
}

//go:inline
func in[LI Value](fieldName string, vals ...any) Query[LI] {
	return func(fi FieldIndexFn[LI], _ *BitSet[LI]) (_ *BitSet[LI], canMutate bool, _ error) {
		if len(vals) == 0 {
			return NewBitSet[LI](), true, nil
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
func NotEq(fieldName string, val any) Query32 {
	return notEq[uint32](fieldName, val)
}

//go:inline
func notEq[LI Value](fieldName string, val any) Query[LI] {
	return func(fi FieldIndexFn[LI], allIDs *BitSet[LI]) (_ *BitSet[LI], canMutate bool, _ error) {
		eq := rel[LI](fieldName, Equal, val)
		return Not(eq)(fi, allIDs)
	}
}

// Not Not(Query)
func Not[LI Value](q Query[LI]) Query[LI] {
	return func(fi FieldIndexFn[LI], allIDs *BitSet[LI]) (_ *BitSet[LI], canMutate bool, _ error) {
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
func And[LI Value](a Query[LI], b Query[LI]) Query[LI] {
	return func(fi FieldIndexFn[LI], allIDs *BitSet[LI]) (_ *BitSet[LI], canMutate bool, _ error) {
		result, err := ensureMutable(a(fi, allIDs))
		if err != nil {
			return nil, false, err
		}
		right, _, err := b(fi, allIDs)
		if err != nil {
			return nil, false, err
		}

		result.And(right)
		return result, true, nil
	}
}

// Or Query or Query
func Or[LI Value](a Query[LI], b Query[LI]) Query[LI] {
	return func(fi FieldIndexFn[LI], allIDs *BitSet[LI]) (_ *BitSet[LI], canMutate bool, _ error) {
		result, err := ensureMutable(a(fi, allIDs))
		if err != nil {
			return nil, false, err
		}
		right, _, err := b(fi, allIDs)
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
func ensureMutable[LI Value](b *BitSet[LI], canMutate bool, err error) (*BitSet[LI], error) {
	if err != nil {
		return nil, err
	}

	if canMutate {
		return b, nil
	}

	return b.Copy(), nil
}
