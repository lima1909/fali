package query

type tokenType int

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
	Type tokenType
	Lit  string
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
		return token{tokEOF, ""}
	}

	ch := l.input[l.pos]

	switch {
	case ch == '(':
		l.pos++
		return token{tokLParen, ""} // No need to store literal "("
	case ch == ')':
		l.pos++
		return token{tokRParen, ""}
	case ch == '=':
		l.pos++
		return token{tokEq, ""}
	case ch == '"', ch == '\'':
		return l.readString(ch)
	case (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_':
		return l.readIdentOrKeyword()
	case (ch >= '0' && ch <= '9') || ch == '-':
		return l.readNumber()
	}

	l.pos++
	return token{tokEOF, ""}
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

	lit := l.input[start:l.pos]

	// Zero-allocation keyword check
	if len(lit) == 3 && (lit[0] == 'a' || lit[0] == 'A') && (lit[1] == 'n' || lit[1] == 'N') && (lit[2] == 'd' || lit[2] == 'D') {
		return token{tokAnd, ""}
	}
	if len(lit) == 2 && (lit[0] == 'o' || lit[0] == 'O') && (lit[1] == 'r' || lit[1] == 'R') {
		return token{tokOr, ""}
	}

	return token{tokIdent, lit}
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
	return token{tokNumber, l.input[start:l.pos]}
}

func (l *lexer) readString(quote byte) token {
	l.pos++ // Skip open quote
	start := l.pos
	for l.pos < len(l.input) && l.input[l.pos] != quote {
		l.pos++
	}
	lit := l.input[start:l.pos]
	l.pos++ // Skip close quote
	return token{tokString, lit}
}
