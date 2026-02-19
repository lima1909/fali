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

type parser struct {
	lex *lexer
	cur token
}

func Parse(input string) (Filter, error) {
	p := &parser{lex: &lexer{input: input, pos: 0}}
	p.next()
	return p.parseOr()
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
	if p.cur.Type == tokLParen {
		p.next()
		expr, err := p.parseOr() // Back to the top of the precedence chain
		if err != nil {
			return nil, err
		}
		if p.cur.Type != tokRParen {
			return nil, fmt.Errorf("expected ')'")
		}
		p.next()
		return expr, nil
	}

	if p.cur.Type != tokIdent {
		return nil, fmt.Errorf("expected field, got '%s'", p.cur.Lit)
	}
	field := p.cur.Lit
	p.next()

	if p.cur.Type != tokEq {
		return nil, fmt.Errorf("expected '='")
	}
	p.next()

	var val any
	switch p.cur.Type {
	case tokString:
		val = p.cur.Lit
	case tokNumber:
		val, _ = strconv.Atoi(p.cur.Lit)
	default:
		return nil, fmt.Errorf("expected value")
	}
	p.next()

	return &Eq{Field: field, Value: val}, nil
}
