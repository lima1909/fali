package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser_Base(t *testing.T) {
	getter := func(data map[string]any) Getter {
		return func(field string) any { return data[field] }
	}

	user := map[string]any{"name": "Alice", "role": "admin", "status": 0, "deleted": 1}

	tests := []struct {
		query    string
		expected bool
	}{
		{query: `role="admin"`, expected: true},
		// (role = admin or status = 1) and delete = 1
		{query: `role = "admin" OR status = 1 AND deleted = 1`, expected: true},
		{query: `role = "admin" OR status = 1 AND deleted = 0`, expected: true},
		{query: `role = "admin" OR (status = 1 AND deleted = 1)`, expected: true},
		{query: `role = "user" OR (status = 0 AND deleted = 1)`, expected: true},
		{query: `status = 1 AND deleted = 1`, expected: false},
		{query: `status = 1 or deleted = 0`, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			// query := `role = "admin" OR status = 1 AND deleted = 0`
			ast, err := Parse(tt.query)
			assert.NoError(t, err)

			match := ast.Match(getter(user))
			assert.Equal(t, tt.expected, match)
		})
	}

}
