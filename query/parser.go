package query

import (
	"fmt"
	"strconv"
)

type Getter func(field string) any

type Filter interface {
	Match(get Getter) bool
}

type Eq struct {
	Field string
	Value any
}

func (e *Eq) Match(get Getter) bool { return get(e.Field) == e.Value }

type Neq struct {
	Field string
	Value any
}

func (n *Neq) Match(get Getter) bool {
	return get(n.Field) != n.Value
}

type And struct {
	Left  Filter
	Right Filter
}

func (a *And) Match(get Getter) bool { return a.Left.Match(get) && a.Right.Match(get) }

type Or struct {
	Left  Filter
	Right Filter
}

func (o *Or) Match(get Getter) bool { return o.Left.Match(get) || o.Right.Match(get) }

type Not struct {
	Expr Filter
}

func (n *Not) Match(get Getter) bool {
	return !n.Expr.Match(get)
}

type parser struct {
	input string
	lex   lexer
	cur   token
}

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

func Parse(input string) (Filter, error) {
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

func (p *parser) parseOr() (Filter, error) {
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
		left = &Or{Left: left, Right: right}
	}
	return left, nil
}

func (p *parser) parseAnd() (Filter, error) {
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
		left = &And{Left: left, Right: right}
	}
	return left, nil
}

func (p *parser) parseCondition() (Filter, error) {
	if p.cur.Type == tokNot {
		p.next() // consume 'NOT'
		// Recursively parse the expression that follows
		expr, err := p.parseCondition()
		if err != nil {
			return nil, err
		}
		return &Not{Expr: expr}, nil
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
	case tokEq, tokNeq:
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
	default:
		return nil, ErrUnexpectedToken{token: p.cur, expected: tokString}
	}
	p.next()

	switch relTokenType {
	case tokNeq:
		return &Neq{Field: field, Value: val}, nil
	default:
		// must be Eq, the evaluation was already
		return &Eq{Field: field, Value: val}, nil
	}
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

	var v int
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return nil, strconv.ErrSyntax
		}
		v = v*10 + int(c-'0')
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
func Optimize(f Filter) Filter {
	switch node := f.(type) {

	case *And:
		node.Left = Optimize(node.Left)
		node.Right = Optimize(node.Right)
		return node

	case *Or:
		node.Left = Optimize(node.Left)
		node.Right = Optimize(node.Right)
		return node

	case *Not:
		// 1. Optimize the inner expression first
		inner := Optimize(node.Expr)

		// 2. Apply our Transformation Rules
		switch in := inner.(type) {
		case *Eq:
			// Rule: NOT (A = B)  -->  A != B
			return &Neq{Field: in.Field, Value: in.Value}

		case *Neq:
			// Rule: NOT (A != B)  -->  A = B
			return &Eq{Field: in.Field, Value: in.Value}

		case *Not:
			// Rule: Double Negation: NOT (NOT A) --> A
			return in.Expr

		// --- NEW: De Morgan's Laws ---
		case *And:
			// Rule: NOT (A AND B) --> (NOT A) OR (NOT B)
			newOr := &Or{
				Left:  &Not{Expr: in.Left},
				Right: &Not{Expr: in.Right},
			}
			// Recursively optimize the new structure to collapse the NOTs!
			return Optimize(newOr)

		case *Or:
			// Rule: NOT (A OR B) --> (NOT A) AND (NOT B)
			newAnd := &And{
				Left:  &Not{Expr: in.Left},
				Right: &Not{Expr: in.Right},
			}
			// Recursively optimize the new structure to collapse the NOTs!
			return Optimize(newAnd)
		// -----------------------------

		default:
			node.Expr = inner
			return node
		}

	default:
		return f
	}
}
