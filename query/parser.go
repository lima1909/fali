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
	input string
	lex   lexer
	cur   token
}

type ParseError struct {
	msg   string
	token token
}

func (e ParseError) Error() string {
	return fmt.Sprintf("%s [%d:%d]", e.msg, e.token.Start, e.token.End)
}

func Parse(input string) (Filter, error) {
	p := parser{input: input, lex: lexer{input: input, pos: 0}}
	p.next()
	ast, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	if p.cur.Type != tokEOF {
		return nil, ParseError{"unexpected token", p.cur}
	}
	return ast, nil
}

//go:inline
func (p *parser) next() {
	p.cur = p.lex.nextToken()
}

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
			return nil, ParseError{"expected ')'", p.cur}
		}
		p.next()
		return expr, nil
	}

	if p.cur.Type != tokIdent {
		return nil, ParseError{"expected field", p.cur}
	}
	field := p.input[p.cur.Start:p.cur.End]
	p.next()

	if p.cur.Type != tokEq {
		return nil, ParseError{"expected relation like: '='", p.cur}
	}
	p.next()

	var val any
	switch p.cur.Type {
	case tokString:
		val = p.input[p.cur.Start:p.cur.End]
	case tokNumber:
		num, err := p.parseNumber()
		if err != nil {
			return nil, ParseError{"expected value", p.cur}
		}
		val = num
	default:
		return nil, ParseError{"expected value", p.cur}
	}
	p.next()

	return &Eq{Field: field, Value: val}, nil
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
