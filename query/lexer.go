package query

type tokenType uint8

const (
	tokEOF tokenType = iota
	tokIdent
	tokString
	tokNumber
	tokEq
	tokAnd
	tokOr
	tokLParen
	tokRParen
)

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
	case ch == '"', ch == '\'':
		return l.readString(ch)
	case (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_':
		return l.readIdentOrKeyword()
	case (ch >= '0' && ch <= '9') || ch == '-':
		return l.readNumber()
	}

	l.pos++
	return token{Type: tokEOF, Start: l.pos, End: l.pos}
}

// readIdentOrKeyword checks if the word is AND / OR without allocating memory
func (l *lexer) readIdentOrKeyword() token {
	start := l.pos
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' {
			l.pos++
		} else {
			break
		}
	}

	length := l.pos - start
	// Zero-allocation keyword check
	if length == 3 {
		b := l.input[start:]
		if (b[0] == 'a' || b[0] == 'A') && (b[1] == 'n' || b[1] == 'N') && (b[2] == 'd' || b[2] == 'D') {
			return token{Type: tokAnd, Start: start, End: l.pos}
		}
	}
	if length == 2 {
		b := l.input[start:]
		if (b[0] == 'o' || b[0] == 'O') && (b[1] == 'r' || b[1] == 'R') {
			return token{Type: tokOr, Start: start, End: l.pos}
		}
	}

	return token{Type: tokIdent, Start: start, End: l.pos}
}

func (l *lexer) readNumber() token {
	start := l.pos
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if (ch >= '0' && ch <= '9') || ch == '.' {
			l.pos++
		} else {
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
