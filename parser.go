package main

import (
	"fmt"
	"math"
	"strconv"
)

type ErrUnexpectedToken struct {
	token    token
	expected tokenType
}

func (e ErrUnexpectedToken) Error() string {
	if e.expected == tokUndefined {
		return fmt.Sprintf(
			"unexpected token: %q [%d:%d]",
			e.token.Type,
			e.token.Start,
			e.token.End,
		)
	}
	return fmt.Sprintf(
		"unexpected token: %q, expected: %q [%d:%d]",
		e.token.Type,
		e.expected,
		e.token.Start,
		e.token.End,
	)
}

type ErrCast struct{ msg string }

func (e ErrCast) Error() string { return fmt.Sprintf("cast err: %s", e.msg) }

type parser struct {
	input string
	lex   lexer
	cur   token
}

func Parse(input string) (Query32, error) {
	p := parser{input: input, lex: lexer{input: input, pos: 0}}
	p.next()
	ast, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	if p.cur.Type != tokEOF {
		return nil, ErrUnexpectedToken{token: p.cur}
	}
	return ast, nil
}

//go:inline
func (p *parser) next() { p.cur = p.lex.nextToken() }

func (p *parser) parseOr() (Query32, error) {
	// the rule: AND before OR
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.cur.Type == tokOr {
		p.next()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = Or(left, right)
	}
	return left, nil
}

func (p *parser) parseAnd() (Query32, error) {
	left, err := p.parseCondition()
	if err != nil {
		return nil, err
	}

	for p.cur.Type == tokAnd {
		p.next()
		right, err := p.parseCondition()
		if err != nil {
			return nil, err
		}
		left = And(left, right)
	}
	return left, nil
}

func (p *parser) parseCondition() (Query32, error) {
	if p.cur.Type == tokNot {
		p.next() // consume 'NOT'
		// Recursively parse the expression that follows
		expr, err := p.parseCondition()
		if err != nil {
			return nil, err
		}
		return Not(expr), nil
	}

	if p.cur.Type == tokLParen {
		p.next()
		expr, err := p.parseOr() // Back to the top of the precedence chain
		if err != nil {
			return nil, err
		}
		if p.cur.Type != tokRParen {
			return nil, ErrUnexpectedToken{token: p.cur, expected: tokRParen}
		}
		p.next()
		return expr, nil
	}

	if p.cur.Type != tokIdent {
		return nil, ErrUnexpectedToken{token: p.cur, expected: tokIdent}
	}
	field := p.input[p.cur.Start:p.cur.End]
	p.next()

	// is the relation supported
	relTokenType := p.cur.Type
	switch relTokenType {
	case tokEq, tokNeq, tokLess, tokLessEq, tokGreater, tokGreaterEq:
	// supported relation, do nothing here
	default:
		return nil, ErrUnexpectedToken{token: p.cur, expected: tokEq}
	}
	p.next()

	var val any
	switch p.cur.Type {
	case tokString:
		val = p.input[p.cur.Start:p.cur.End]
	case tokNumber:
		num, err := p.parseNumber()
		if err != nil {
			return nil, ErrUnexpectedToken{token: p.cur, expected: tokNumber}
		}
		val = num
	case tokBool:
		boolean, err := strconv.ParseBool(p.input[p.cur.Start:p.cur.End])
		if err != nil {
			return nil, ErrUnexpectedToken{token: p.cur, expected: tokBool}
		}
		val = boolean

	// value with cast: uint8(10)
	case tokIdent:
		typeName := p.input[p.cur.Start:p.cur.End]
		p.next()
		if p.cur.Type != tokLParen {
			return nil, ErrUnexpectedToken{token: p.cur, expected: tokLParen}
		}
		p.next()

		if p.cur.Type == tokNumber {
			num, err := p.parseNumber()
			if err != nil {
				return nil, ErrUnexpectedToken{token: p.cur, expected: tokNumber}
			}
			val, err = castValue(typeName, num)
			if err != nil {
				return nil, err
			}
			p.next()
		}

		if p.cur.Type != tokRParen {
			return nil, ErrUnexpectedToken{token: p.cur, expected: tokRParen}
		}
	default:
		return nil, ErrUnexpectedToken{token: p.cur, expected: tokString}
	}
	p.next()

	switch relTokenType {
	case tokNeq:
		return NotEq(field, val), nil
	case tokLess:
		return Lt(field, val), nil
	case tokLessEq:
		return Le(field, val), nil
	case tokGreater:
		return Gt(field, val), nil
	case tokGreaterEq:
		return Ge(field, val), nil
	default:
		// must be Eq, the evaluation was already
		return Eq(field, val), nil
	}
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

// Optimize rewrites the AST to eliminate unnecessary nesting and function calls.
// Apply the AST transformations!
// optimizedAst := Optimize(ast)
//
// return optimizedAst, nil
// func Optimize(f Filter) Filter {
// 	switch node := f.(type) {
//
// 	case *And:
// 		node.Left = Optimize(node.Left)
// 		node.Right = Optimize(node.Right)
// 		return node
//
// 	case *Or:
// 		node.Left = Optimize(node.Left)
// 		node.Right = Optimize(node.Right)
// 		return node
//
// 	case *Not:
// 		// 1. Optimize the inner expression first
// 		inner := Optimize(node.Expr)
//
// 		// 2. Apply our Transformation Rules
// 		switch in := inner.(type) {
// 		case *Eq:
// 			// Rule: NOT (A = B)  -->  A != B
// 			return &Neq{Field: in.Field, Value: in.Value}
//
// 		case *Neq:
// 			// Rule: NOT (A != B)  -->  A = B
// 			return &Eq{Field: in.Field, Value: in.Value}
//
// 		case *Not:
// 			// Rule: Double Negation: NOT (NOT A) --> A
// 			return in.Expr
//
// 		// --- NEW: De Morgan's Laws ---
// 		case *And:
// 			// Rule: NOT (A AND B) --> (NOT A) OR (NOT B)
// 			newOr := &Or{
// 				Left:  &Not{Expr: in.Left},
// 				Right: &Not{Expr: in.Right},
// 			}
// 			// Recursively optimize the new structure to collapse the NOTs!
// 			return Optimize(newOr)
//
// 		case *Or:
// 			// Rule: NOT (A OR B) --> (NOT A) AND (NOT B)
// 			newAnd := &And{
// 				Left:  &Not{Expr: in.Left},
// 				Right: &Not{Expr: in.Right},
// 			}
// 			// Recursively optimize the new structure to collapse the NOTs!
// 			return Optimize(newAnd)
// 		// -----------------------------
//
// 		default:
// 			node.Expr = inner
// 			return node
// 		}
//
// 	default:
// 		return f
// 	}
// }
