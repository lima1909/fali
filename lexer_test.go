package main

import (
	"fmt"
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
		{query: `-9`, expected: tokNumber},
		{query: `-0.9`, expected: tokNumber},
		{query: `"false"`, expected: tokString},
		{query: `Or`, expected: tokOr},
		{query: `aND`, expected: tokAnd},
		{query: ` noT `, expected: tokNot},
		{query: ` = `, expected: tokEq},
		{query: ` != `, expected: tokNeq},
		{query: ` < `, expected: tokLess},
		{query: `<=`, expected: tokLessEq},
		{query: ` > `, expected: tokGreater},
		{query: `>=`, expected: tokGreaterEq},
		{query: `(`, expected: tokLParen},
		{query: `)`, expected: tokRParen},

		{query: `startswith`, expected: tokIdent},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			lex := lexer{input: tt.query, pos: 0}
			lexerToken := lex.nextToken().Type
			assert.Equal(
				t,
				tt.expected,
				lexerToken,
				fmt.Sprintf("%s != %s", tt.expected, lexerToken),
			)
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
		{query: `num = -5`, expected: []tokenType{
			tokIdent,
			tokEq,
			tokNumber,
		}},
		{query: `num = -5.3`, expected: []tokenType{
			tokIdent,
			tokEq,
			tokNumber,
		}},
		{query: `float32(-5)`, expected: []tokenType{
			tokIdent,
			tokLParen,
			tokNumber,
			tokRParen,
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

		{query: `name startswith "Ma"`, expected: []tokenType{
			tokIdent,
			tokIdent,
			tokString,
		}},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			lex := lexer{input: tt.query, pos: 0}
			for _, token := range tt.expected {
				lexerToken := lex.nextToken().Type
				assert.Equal(
					t,
					token,
					lexerToken,
					fmt.Sprintf("%s != %s", token, lexerToken),
				)
			}
		})
	}
}

func TestLexer_Invalid(t *testing.T) {

	tests := []struct {
		query    string
		expected []tokenType
	}{
		{query: `3.3.1`, expected: []tokenType{
			tokNumber,
			tokEOF,
		}},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			lex := lexer{input: tt.query, pos: 0}
			for _, token := range tt.expected {
				lexerToken := lex.nextToken().Type
				assert.Equal(
					t,
					token,
					lexerToken,
					fmt.Sprintf("%s != %s", token, lexerToken),
				)
			}
		})
	}
}
