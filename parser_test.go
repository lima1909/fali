package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type User struct {
	name  string
	role  string
	ok    bool
	price float64
}

func (u *User) Name() string   { return u.name }
func (u *User) Role() string   { return u.role }
func (u *User) Ok() bool       { return u.ok }
func (u *User) Price() float64 { return u.price }

func TestParser_Base(t *testing.T) {
	user := User{name: "Alice", role: "admin", ok: false, price: 1.2}

	indexMap := newIndexMap[User, struct{}](nil)
	indexMap.index["name"] = NewSortedIndex((*User).Name)
	indexMap.index["name"].Set(&user, 1)
	indexMap.index["role"] = NewSortedIndex((*User).Role)
	indexMap.index["role"].Set(&user, 1)
	indexMap.index["price"] = NewMapIndex((*User).Price)
	// indexMap.index["price"].Set(&User{}, 0)
	indexMap.index["price"].Set(&user, 1)
	indexMap.index["ok"] = NewMapIndex((*User).Ok)
	indexMap.index["ok"].Set(&User{ok: true}, 0)
	indexMap.index["ok"].Set(&user, 1)
	indexMap.allIDs.Set(0)
	indexMap.allIDs.Set(1)

	// the rule: AND before OR
	tests := []struct {
		query    string
		expected []uint32
	}{
		{query: `role="admin"`, expected: []uint32{1}},
		{query: `price = 1.2`, expected: []uint32{1}},
		{query: `price = 4.2`, expected: []uint32{}},
		{query: `ok = false`, expected: []uint32{1}},
		{query: `ok = true`, expected: []uint32{0}},
		{query: `NOT(ok = true)`, expected: []uint32{1}},

		{query: `ok = true or price = 0.0`, expected: []uint32{0}},

		{query: `role = "admin" AND price = 9.9`, expected: []uint32{}},
		{query: `role = "admin" OR price = 9.9`, expected: []uint32{1}},
		{query: `not (ok = true or price = 0.0)`, expected: []uint32{1}},

		//  true or (false and true) => true
		{query: `role = "admin" OR ok = true AND price = 1.2`, expected: []uint32{1}},
		// true or (false and false) => true
		{query: `role = "admin" OR ok = true AND price = 0.0`, expected: []uint32{1}},
		// true or (true and true) => true
		{query: `role = "admin" OR (ok = true AND price = 1.2)`, expected: []uint32{1}},
		// false or (true and true) => true
		{query: `role = "user" OR (ok = false AND price = 1.2)`, expected: []uint32{1}},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			query, err := Parse(tt.query)
			assert.NoError(t, err)

			bs, _, err := query(indexMap.LookupByName, indexMap.allIDs)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, bs.ToSlice())
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
