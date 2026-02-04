package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	// "github.com/stretchr/testify/assert"
)

type car struct {
	name  string
	color string
	age   uint8
	isNew bool
}

func TestIndexList_Base(t *testing.T) {
	il := NewIndexList[car]()
	il.fieldIndexMap = FieldIndexMap[car, uint32]{
		"name":  {index: NewMapIndex[uint32](), fieldFn: func(c *car) any { return c.name }},
		"age":   {index: NewMapIndex[uint32](), fieldFn: func(c *car) any { return c.age }},
		"isnew": {index: NewMapIndex[uint32](), fieldFn: func(c *car) any { return c.isNew }},
	}

	il.Add(car{name: "Dacia", age: 22, color: "red"})
	il.Add(car{name: "Opel", age: 22})
	il.Add(car{name: "Mercedes", age: 5, isNew: true})
	il.Add(car{name: "Dacia", age: 22})

	count, err := il.Query(Eq[uint32]("name", "Dacia"))
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	count, err = il.Query(Eq[uint32]("name", "Opel"))
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	count, err = il.Query(Eq[uint32]("age", uint8(5)))
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	count, err = il.Query(Eq[uint32]("isnew", false))
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
	count, err = il.Query(Eq[uint32]("isnew", true))
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// wrong value type, expected: uint8, got int
	count, err = il.Query(Eq[uint32]("age", 5))
	assert.Error(t, err)
	assert.Equal(t, 0, count)

	// wrong field name, expected: age, got wrong
	count, err = il.Query(Eq[uint32]("wrong", 5))
	assert.Error(t, err)
	assert.Equal(t, 0, count)
}
