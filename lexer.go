package main

import "fmt"

type tokenType uint8

const (
	tokUndefined tokenType = iota
	// end of file
	tokEOF
	// ident
	tokIdent
	// datatypes
	tokString
	tokNumber
	tokBool
	// relations
	tokEq
	tokNeq
	tokLess
	tokLessEq
	tokGreater
	tokGreaterEq
	// logical combinations
	tokAnd
	tokOr
	tokNot
	// parentheses
	tokLParen
	tokRParen
)

func (t tokenType) String() string {
	switch t {
	case tokUndefined:
		return "undefined"
	case tokEOF:
		return "EOF"
	case tokIdent:
		return "indent"
	case tokString:
		return "string"
	case tokNumber:
		return "number"
	case tokBool:
		return "bool"
	case tokEq:
		return "="
	case tokNeq:
		return "!="
	case tokLess:
		return "<"
	case tokLessEq:
		return "<="
	case tokGreater:
		return ">"
	case tokGreaterEq:
		return ">="
	case tokAnd:
		return "and"
	case tokOr:
		return "or"
	case tokNot:
		return "not"
	case tokLParen:
		return "("
	case tokRParen:
		return ")"
	default:
		return fmt.Sprintf("UNKNOWN: %d", t)
	}
}

type token struct {
	Start int
	End   int
	Type  tokenType
}

type lexer struct {
	input string
	pos   int
}

func (l *lexer) nextToken() token {
	// skip whitespace
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			l.pos++
			continue
		}
		break
	}

	if l.pos >= len(l.input) {
		return token{Type: tokEOF, Start: l.pos, End: l.pos}
	}

	ch := l.input[l.pos]

	switch {
	case ch == '(':
		start := l.pos
		l.pos++
		return token{Type: tokLParen, Start: start, End: l.pos}
	case ch == ')':
		start := l.pos
		l.pos++
		return token{Type: tokRParen, Start: start, End: l.pos}
	case ch == '=':
		start := l.pos
		l.pos++
		return token{Type: tokEq, Start: start, End: l.pos}

	case ch == '!':
		start := l.pos
		// Check if the next byte exists and is '='
		if l.pos+1 < len(l.input) && l.input[l.pos+1] == '=' {
			l.pos += 2 // Consume both '!' and '='
			return token{Type: tokNeq, Start: start, End: l.pos}
		}
		// Optional: Handle a lone '!' if you want a NOT operator later
		l.pos++
	case ch == '<':
		start := l.pos
		if l.pos+1 < len(l.input) && l.input[l.pos+1] == '=' {
			l.pos += 2
			return token{Type: tokLessEq, Start: start, End: l.pos}
		}
		l.pos++
		return token{Type: tokLess, Start: start, End: l.pos}
	case ch == '>':
		start := l.pos
		if l.pos+1 < len(l.input) && l.input[l.pos+1] == '=' {
			l.pos += 2
			return token{Type: tokGreaterEq, Start: start, End: l.pos}
		}
		l.pos++
		return token{Type: tokGreater, Start: start, End: l.pos}
	case ch == '"', ch == '\'':
		return l.readString(ch)
	case (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_':
		return l.readBoolOrIdentOrKeyword()
	case (ch >= '0' && ch <= '9') || ch == '-':
		return l.readNumber()
	}

	l.pos++
	return token{Type: tokEOF, Start: l.pos, End: l.pos}
}

// readIdentOrKeyword checks if the word is AND / OR without allocating memory
func (l *lexer) readBoolOrIdentOrKeyword() token {
	start := l.pos
	// read while are there letters, numbers or _
	// it starts with a letter
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' {
			l.pos++
		} else {
			break
		}
	}

	length := l.pos - start
	b := l.input[start:]

	// evaluate the founded string
	// Fast Keyword & Boolean Checks
	switch length {
	case 2:
		if (b[0] == 'o' || b[0] == 'O') && (b[1] == 'r' || b[1] == 'R') {
			return token{Type: tokOr, Start: start, End: l.pos}
		}
	case 3:
		if (b[0] == 'a' || b[0] == 'A') && (b[1] == 'n' || b[1] == 'N') && (b[2] == 'd' || b[2] == 'D') {
			return token{Type: tokAnd, Start: start, End: l.pos}
		}
		if (b[0] == 'n' || b[0] == 'N') && (b[1] == 'o' || b[1] == 'O') && (b[2] == 't' || b[2] == 'T') {
			return token{Type: tokNot, Start: start, End: l.pos}
		}
	case 4:
		// check for "true" (Case Insensitive)
		if (b[0] == 't' || b[0] == 'T') &&
			(b[1] == 'r' || b[1] == 'R') &&
			(b[2] == 'u' || b[2] == 'U') &&
			(b[3] == 'e' || b[3] == 'E') {
			return token{Type: tokBool, Start: start, End: l.pos}
		}
	case 5:
		// check for "false" (Case Insensitive)
		if (b[0] == 'f' || b[0] == 'F') &&
			(b[1] == 'a' || b[1] == 'A') &&
			(b[2] == 'l' || b[2] == 'L') &&
			(b[3] == 's' || b[3] == 'S') &&
			(b[4] == 'e' || b[4] == 'E') {
			return token{Type: tokBool, Start: start, End: l.pos}
		}
	}

	// If it didn't match any of the keywords, it's just a normal identifier
	return token{Type: tokIdent, Start: start, End: l.pos}
}

func (l *lexer) readNumber() token {
	start := l.pos
	hasDot := false

	if l.pos < len(l.input) && l.input[l.pos] == '-' {
		l.pos++
	}

	for l.pos < len(l.input) {
		ch := l.input[l.pos]

		if ch >= '0' && ch <= '9' {
			l.pos++
		} else if ch == '.' && !hasDot {
			// First time seeing a dot, mark it and continue
			hasDot = true
			l.pos++
		} else {
			// If it's a second dot, a letter, or a space, we are done!
			break
		}
	}

	return token{Type: tokNumber, Start: start, End: l.pos}
}

func (l *lexer) readString(quote byte) token {
	l.pos++ // Skip open quote
	start := l.pos
	for l.pos < len(l.input) && l.input[l.pos] != quote {
		l.pos++
	}
	end := l.pos
	if l.pos < len(l.input) {
		l.pos++ // Skip close quote
	}
	return token{Type: tokString, Start: start, End: end}
}
