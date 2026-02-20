package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser_Base(t *testing.T) {
	getter := func(data map[string]any) Getter {
		return func(field string) any { return data[field] }
	}

	user := map[string]any{"name": "Alice", "role": "admin", "status": 0, "price": 1.2}

	tests := []struct {
		query    string
		expected bool
	}{
		{query: `role="admin"`, expected: true},
		{query: `price = 1.2`, expected: true},
		{query: `status = 0`, expected: true},
		// (role = admin or status = 1) and delete = 1
		{query: `role = "admin" OR status = 1 AND price = 1.2`, expected: true},
		{query: `role = "admin" OR status = 1 AND price = 0`, expected: true},
		{query: `role = "admin" OR (status = 1 AND price = 1.2)`, expected: true},
		{query: `role = "user" OR (status = 0 AND price = 1.2)`, expected: true},
		{query: `status = 1 AND price = 1.2`, expected: false},
		{query: `status = 1 or price = 0`, expected: false},
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
		query    string
		expected string
	}{
		{query: ``, expected: "expected field"},
		{query: `role`, expected: "expected relation"},
		{query: `role ~`, expected: "expected relation"},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			_, err := Parse(tt.query)
			assert.Contains(t, err.Error(), tt.expected)
			// fmt.Println("--->>>", err)
		})
	}

}
