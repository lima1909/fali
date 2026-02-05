package main

// FieldIndexFn is the interfact to the FieldIndexMap
type FieldIndexFn[R Row] = func(string, any) (Index[R], error)

// Query is a filter function, find the correct Index an execute the Index.Get method
// and returns a BitSet pointer
type Query[R Row] func(fi FieldIndexFn[R], allIDs *BitSet[R]) (_ *BitSet[R], canMutate bool, _ error)

// All means returns all Items, no filtering
func All[R Row]() Query[R] {
	return func(_ FieldIndexFn[R], allIDs *BitSet[R]) (_ *BitSet[R], canMutate bool, _ error) {
		return allIDs, false, nil
	}
}

// Eq fieldName = val
func Eq[R Row](fieldName string, val any) Query[R] {
	return func(fi FieldIndexFn[R], _ *BitSet[R]) (_ *BitSet[R], canMutate bool, _ error) {
		idx, err := fi(fieldName, val)
		if err != nil {
			return nil, false, err
		}

		return idx.Get(Equal, val), false, nil
	}
}

// In combines Eq with an Or
// In("name", "Paul", "Egon") => name == "Paul" Or name == "Egon"
func In[R Row](fieldName string, vals ...any) Query[R] {
	return func(fi FieldIndexFn[R], _ *BitSet[R]) (_ *BitSet[R], canMutate bool, _ error) {
		if len(vals) == 0 {
			return NewBitSet[R](), true, nil
		}

		idx, err := fi(fieldName, vals[0])
		if err != nil {
			return nil, false, err
		}

		bs := idx.Get(Equal, vals[0])
		if len(vals) == 1 {
			return bs, false, nil
		}

		bs = bs.Copy()
		for _, val := range vals[1:] {
			bs.Or(idx.Get(Equal, val))
		}

		return bs, true, nil
	}
}

// NotEq is a shorcut for Not(Eq(...))
func NotEq[R Row](fieldName string, val any) Query[R] {
	return func(fi FieldIndexFn[R], allIDs *BitSet[R]) (_ *BitSet[R], canMutate bool, _ error) {
		eq := Eq[R](fieldName, val)
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
