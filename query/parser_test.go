package query

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLexer(t *testing.T) {

	tests := []struct {
		query    string
		expected tokenType
	}{
		{query: `true`, expected: tokBool},
		{query: `tRue`, expected: tokBool},
		{query: `fALse`, expected: tokBool},
		{query: `4.2`, expected: tokNumber},
		{query: `7`, expected: tokNumber},
		{query: `0.9`, expected: tokNumber},
		{query: `"false"`, expected: tokString},
		{query: `Or`, expected: tokOr},
		{query: `aND`, expected: tokAnd},
		{query: ` noT `, expected: tokNot},
		{query: ` = `, expected: tokEq},
		{query: ` != `, expected: tokNeq},
		{query: `(`, expected: tokLParen},
		{query: `)`, expected: tokRParen},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			lex := lexer{input: tt.query, pos: 0}
			assert.Equal(t, tt.expected, lex.nextToken().Type)
		})
	}
}

func TestParser_Base(t *testing.T) {
	getter := func(data map[string]any) Getter {
		return func(field string) any { return data[field] }
	}

	user := map[string]any{"name": "Alice", "role": "admin", "ok": false, "price": 1.2}

	tests := []struct {
		query    string
		expected bool
	}{
		{query: `role="admin"`, expected: true},
		{query: `price = 1.2`, expected: true},
		{query: `ok = false`, expected: true},
		{query: `NOT(ok = true)`, expected: true},
		{query: `role = "admin" OR ok = true AND price = 1.2`, expected: true},
		{query: `role = "admin" OR ok = true AND price = 0`, expected: true},
		{query: `role = "admin" OR (ok = true AND price = 1.2)`, expected: true},
		{query: `role = "user" OR (ok = false AND price = 1.2)`, expected: true},
		{query: `status = 1 AND price = 1.2`, expected: false},
		{query: `status = 1 or price = 0`, expected: false},
		{query: `not (status = 1 or price = 0)`, expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			ast, err := Parse(tt.query)
			assert.NoError(t, err)

			match := ast.Match(getter(user))
			assert.Equal(t, tt.expected, match)
		})
	}

}

func TestParser_Error(t *testing.T) {

	tests := []struct {
		query              string
		expected_tokentype tokenType
		err_tokentpye      tokenType
	}{
		{
			query:              ``,
			expected_tokentype: tokIdent,
			err_tokentpye:      tokEOF,
		},
		{
			query:              `role`,
			expected_tokentype: tokEq,
			err_tokentpye:      tokEOF,
		},
		{
			query:              `role ~`,
			expected_tokentype: tokEq,
			err_tokentpye:      tokEOF,
		},
		{
			query:              `false`,
			expected_tokentype: tokIdent,
			err_tokentpye:      tokBool,
		},
		{
			query:              `role = `,
			expected_tokentype: tokString,
			err_tokentpye:      tokEOF,
		},
		{
			query:              `(role = 3`,
			expected_tokentype: tokRParen,
			err_tokentpye:      tokEOF,
		},
		{
			query:              `role = 3   and `,
			expected_tokentype: tokIdent,
			err_tokentpye:      tokEOF,
		},
		{
			query:              `role = 3   and 5 `,
			expected_tokentype: tokIdent,
			err_tokentpye:      tokNumber,
		},
		{
			query:              `not 3 `,
			expected_tokentype: tokIdent,
			err_tokentpye:      tokNumber,
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			_, err := Parse(tt.query)
			var parseErr ErrUnexpectedToken
			assert.True(t, errors.As(err, &parseErr))
			assert.Equal(t, tt.err_tokentpye, parseErr.token.Type)
			assert.Equal(t, tt.expected_tokentype, parseErr.expected)
		})
	}
}
