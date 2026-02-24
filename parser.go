package main

import (
	"fmt"
	"math"
	"strconv"
)

type ExprKind uint8

const (
	ExprTerm ExprKind = iota
	ExprOr
	ExprAnd
	ExprAndNot
	ExprNot
)

type Expr interface{ kind() ExprKind }

type BinaryExpr struct {
	Op    ExprKind // exprOr, exprAnd, exprAndNot
	Left  Expr
	Right Expr
}

func (e BinaryExpr) kind() ExprKind { return e.Op }

type NotExpr struct{ Child Expr }

func (e NotExpr) kind() ExprKind { return ExprNot }

type TermExpr struct {
	Field string
	Op    Op
	Value any
}

func (e TermExpr) kind() ExprKind { return ExprTerm }

// Parser impl starts
type parser struct {
	input string
	lex   lexer
	cur   token
}

func optimize(e Expr) Expr {
	if e == nil {
		return nil
	}

	switch n := e.(type) {
	case BinaryExpr:
		left := optimize(n.Left)
		right := optimize(n.Right)

		// RULE: And(A, Not(B)) -> AndNot(A, B)
		if n.Op == ExprAnd {
			if notNode, ok := right.(NotExpr); ok {
				return BinaryExpr{Op: ExprAndNot, Left: left, Right: notNode.Child}
			}
			if notNode, ok := left.(NotExpr); ok {
				return BinaryExpr{Op: ExprAndNot, Left: right, Right: notNode.Child}
			}
		}
		return BinaryExpr{Op: n.Op, Left: left, Right: right}

	case NotExpr:
		child := optimize(n.Child)
		switch c := child.(type) {
		// RULE: Not(Not(A)) -> A (Double Negative)
		case NotExpr:
			return optimize(c.Child)
		case TermExpr:
			switch c.Op {
			// I'm not sure, that this is faster
			// RULE: NOT (A = B)  -->  A != B
			// 	case OpEq:
			// 		return TermExpr{Field: c.Field, Op: OpNeq, Value: c.Value}
			// RULE: NOT (A != B)  -->  A = B
			case OpNeq:
				return TermExpr{Field: c.Field, Op: OpEq, Value: c.Value}
			// RULE: NOT (A > B) --> A <= B
			case OpGt:
				return TermExpr{Field: c.Field, Op: OpLe, Value: c.Value}
			// RULE: NOT (A >= B) --> A < B
			case OpGe:
				return TermExpr{Field: c.Field, Op: OpLt, Value: c.Value}
			// RULE: NOT (A < B) --> A >= B
			case OpLt:
				return TermExpr{Field: c.Field, Op: OpGe, Value: c.Value}
			// RULE: NOT (A <= B) --> A > B
			case OpLe:
				return TermExpr{Field: c.Field, Op: OpGt, Value: c.Value}

			default:
				//  no otimizations
				return n
			}
		default:
			//  no otimizations
			return n

		}
	default:
		return e
	}
}

func compile(e Expr) Query32 {
	switch n := e.(type) {
	case TermExpr:
		return rel[uint32](n.Field, n.Op, n.Value)

	case NotExpr:
		return Not(compile(n.Child))

	case BinaryExpr:
		left := compile(n.Left)
		right := compile(n.Right)

		switch n.Op {
		case ExprAnd:
			return And(left, right)
		case ExprOr:
			return Or(left, right)
		case ExprAndNot:
			// This calls the new high-performance AndNot we discussed!
			return AndNot(left, right)
		}
	}

	return All()
}

func Parse(input string) (Query32, error) {
	p := parser{input: input, lex: lexer{input: input, pos: 0}}
	p.next()
	ast, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	if p.cur.Op != OpEOF {
		return nil, ErrUnexpectedToken{token: p.cur}
	}

	optAst := optimize(ast)
	query := compile(optAst)

	return query, nil
}

//go:inline
func (p *parser) next() { p.cur = p.lex.nextToken() }

func (p *parser) parseOr() (Expr, error) {
	// the rule: AND before OR
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.cur.Op == OpOr {
		p.next()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = BinaryExpr{Op: ExprOr, Left: left, Right: right}
	}
	return left, nil
}

func (p *parser) parseAnd() (Expr, error) {
	left, err := p.parseCondition()
	if err != nil {
		return nil, err
	}

	for p.cur.Op == OpAnd {
		p.next()
		right, err := p.parseCondition()
		if err != nil {
			return nil, err
		}
		left = BinaryExpr{Op: ExprAnd, Left: left, Right: right}
	}
	return left, nil
}

func (p *parser) parseCondition() (Expr, error) {
	if p.cur.Op == OpNot {
		p.next() // consume 'NOT'
		// Recursively parse the expression that follows
		expr, err := p.parseCondition()
		if err != nil {
			return nil, err
		}
		return NotExpr{Child: expr}, nil
	}

	if p.cur.Op == OpLParen {
		p.next()
		expr, err := p.parseOr() // Back to the top of the precedence chain
		if err != nil {
			return nil, err
		}
		if p.cur.Op != OpRParen {
			return nil, ErrUnexpectedToken{token: p.cur, expected: OpRParen}
		}
		p.next()
		return expr, nil
	}

	if p.cur.Op != OpIdent {
		return nil, ErrUnexpectedToken{token: p.cur, expected: OpIdent}
	}
	field := p.input[p.cur.Start:p.cur.End]
	p.next()

	// is the relation supported
	tokenOp := p.cur.Op
	var termExpr TermExpr
	switch tokenOp {
	case OpNeq:
		// make it later to an Neq, if we have the value
		termExpr = TermExpr{Field: field, Op: OpEq}
	case OpLt, OpLe, OpGt, OpGe, OpEq:
		termExpr = TermExpr{Field: field, Op: tokenOp}
	// case tokIdent:
	// maybe relations like startswith
	default:
		return nil, ErrUnexpectedToken{token: p.cur, expected: OpEq}
	}
	p.next()

	var val any
	switch p.cur.Op {
	case OpString:
		val = p.input[p.cur.Start:p.cur.End]
	case OpNumber:
		num, err := p.parseNumber()
		if err != nil {
			return nil, ErrUnexpectedToken{token: p.cur, expected: OpNumber}
		}
		val = num
	case OpBool:
		boolean, err := strconv.ParseBool(p.input[p.cur.Start:p.cur.End])
		if err != nil {
			return nil, ErrUnexpectedToken{token: p.cur, expected: OpBool}
		}
		val = boolean

	// value with cast: uint8(10)
	case OpIdent:
		typeName := p.input[p.cur.Start:p.cur.End]
		p.next()
		if p.cur.Op != OpLParen {
			return nil, ErrUnexpectedToken{token: p.cur, expected: OpLParen}
		}
		p.next()

		if p.cur.Op == OpNumber {
			num, err := p.parseNumber()
			if err != nil {
				return nil, ErrUnexpectedToken{token: p.cur, expected: OpNumber}
			}
			val, err = castValue(typeName, num)
			if err != nil {
				return nil, err
			}
			p.next()
		}

		if p.cur.Op != OpRParen {
			return nil, ErrUnexpectedToken{token: p.cur, expected: OpRParen}
		}
	default:
		return nil, ErrUnexpectedToken{token: p.cur, expected: OpString}
	}
	p.next()

	// set the value for the TermExpr
	termExpr.Value = val
	if tokenOp == OpNeq {
		return NotExpr{Child: termExpr}, nil
	}
	return termExpr, nil
}

func castValue(typeName string, val any) (any, error) {
	switch typeName {
	case "int":
		if v, ok := val.(int64); ok {
			if v < math.MinInt && v > math.MaxInt {
				return nil, ErrCast{"to big for " + typeName}
			}
			return int(v), nil
		}
	case "int8":
		if v, ok := val.(int64); ok {
			if v < math.MinInt8 && v > math.MaxInt8 {
				return nil, ErrCast{"to big for " + typeName}
			}
			return int8(v), nil
		}
	case "int16":
		if v, ok := val.(int64); ok {
			if v < math.MinInt16 && v > math.MaxInt16 {
				return nil, ErrCast{"to big for " + typeName}
			}
			return int16(v), nil
		}
	case "int32":
		if v, ok := val.(int64); ok {
			if v < math.MinInt32 && v > math.MaxInt32 {
				return nil, ErrCast{"to big for " + typeName}
			}
			return int32(v), nil
		}

	case "uint":
		if v, ok := val.(int64); ok {
			if v < 0 || v > math.MaxUint32 {
				return nil, ErrCast{"to big for " + typeName}
			}
			return uint(v), nil
		}
	case "uint8":
		if v, ok := val.(int64); ok {
			if v < 0 || v > math.MaxUint8 {
				return nil, ErrCast{"to big for " + typeName}
			}
			return uint8(v), nil
		}
	case "uint16":
		if v, ok := val.(int64); ok {
			if v < 0 || v > math.MaxUint16 {
				return nil, ErrCast{"to big for " + typeName}
			}
			return uint16(v), nil
		}
	case "uint32":
		if v, ok := val.(int64); ok {
			if v < 0 || v > math.MaxUint32 {
				return nil, ErrCast{"to big for " + typeName}
			}
			return uint32(v), nil
		}

	case "float32":
		if v, ok := val.(int64); ok {
			if float64(v) < math.SmallestNonzeroFloat32 && float64(v) > math.MaxFloat32 {
				return nil, ErrCast{"to big for " + typeName}
			}
			return float32(v), nil
		}
		return float32(val.(float64)), nil
	case "float64":
		if v, ok := val.(int64); ok {
			return float64(v), nil
		}
		return val.(float64), nil
	}

	return nil, fmt.Errorf("unsupported type hint: %s", typeName)
}

func (p *parser) parseNumber() (any, error) {
	s := p.input[p.cur.Start:p.cur.End]
	if len(s) == 0 {
		return nil, strconv.ErrSyntax
	}

	hasDot := false
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			hasDot = true
			break
		}
	}

	if hasDot {
		return strconv.ParseFloat(s, 64)
	}

	negative := false
	i := 0
	if s[0] == '-' {
		negative = true
		i = 1
		if len(s) == 1 {
			return nil, strconv.ErrSyntax
		}
	}

	var v int64
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return nil, strconv.ErrSyntax
		}
		v = v*10 + int64(c-'0')
	}

	if negative {
		v = -v
	}
	return v, nil
}

type ErrUnexpectedToken struct {
	token    token
	expected Op
}

func (e ErrUnexpectedToken) Error() string {
	if e.expected == OpUndefined {
		return fmt.Sprintf(
			"unexpected token: %q [%d:%d]",
			e.token.Op,
			e.token.Start,
			e.token.End,
		)
	}
	return fmt.Sprintf(
		"unexpected token: %q, expected: %q [%d:%d]",
		e.token.Op,
		e.expected,
		e.token.Start,
		e.token.End,
	)
}

type ErrCast struct{ msg string }

func (e ErrCast) Error() string { return fmt.Sprintf("cast err: %s", e.msg) }
