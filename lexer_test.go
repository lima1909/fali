package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLexer_OneToken(t *testing.T) {

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

func TestLexer_ManyToken(t *testing.T) {

	tests := []struct {
		query    string
		expected []tokenType
	}{
		{query: `ok = true`, expected: []tokenType{
			tokIdent,
			tokEq,
			tokBool,
		}},
		{query: `not(ok = true)`, expected: []tokenType{
			tokNot,
			tokLParen,
			tokIdent,
			tokEq,
			tokBool,
			tokRParen,
		}},
		{query: `ok != true`, expected: []tokenType{
			tokIdent,
			tokNeq,
			tokBool,
		}},
		{query: `name = "Inge" and age = 3`, expected: []tokenType{
			tokIdent,
			tokEq,
			tokString,
			tokAnd,
			tokIdent,
			tokEq,
			tokNumber,
		}},
		{query: `name="Inge" or age=3`, expected: []tokenType{
			tokIdent,
			tokEq,
			tokString,
			tokOr,
			tokIdent,
			tokEq,
			tokNumber,
		}},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			lex := lexer{input: tt.query, pos: 0}
			for _, token := range tt.expected {
				assert.Equal(t, token, lex.nextToken().Type)
			}
		})
	}
}
