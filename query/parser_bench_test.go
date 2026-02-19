package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkLexer(b *testing.B) {

	for b.Loop() {
		l := &lexer{input: `role = "admin" OR status = 1 AND deleted = 1`, pos: 0}
		for l.nextToken().Type != tokEOF {
		}
	}
}

func BenchmarkParser(b *testing.B) {
	user := map[string]any{"name": "Alice", "role": "admin", "status": 0, "deleted": 1}
	getter := func(data map[string]any) Getter {
		return func(field string) any { return data[field] }
	}
	b.ResetTimer()

	for b.Loop() {
		ast, err := Parse(`role = "admin" OR status = 1 AND deleted = 1`)
		assert.NoError(b, err)

		match := ast.Match(getter(user))
		assert.True(b, match)
	}
}
