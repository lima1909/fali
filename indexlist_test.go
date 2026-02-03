package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type car struct {
	name  string
	color string
	age   uint8
	isNew bool
}

func TestIndexList_Base(t *testing.T) {
	il := NewIndexList[car]()
	il.fieldIndexMap = map[string]FieldIndex[car, uint32]{
		"name":  {index: NewMapIndex[uint32](), fieldFn: func(c *car) any { return c.name }},
		"age":   {index: NewMapIndex[uint32](), fieldFn: func(c *car) any { return c.age }},
		"isnew": {index: NewMapIndex[uint32](), fieldFn: func(c *car) any { return c.isNew }},
	}

	il.Add(car{name: "Dacia", age: 22, color: "red"})
	il.Add(car{name: "Opel", age: 22})
	il.Add(car{name: "Mercedes", age: 5, isNew: true})
	il.Add(car{name: "Dacia", age: 22})

	assert.Equal(t, 2, il.Query(Eq[uint32]("name", "Dacia")))
	assert.Equal(t, 1, il.Query(Eq[uint32]("name", "Opel")))

	assert.Equal(t, 1, il.Query(Eq[uint32]("age", uint8(5))))

	assert.Equal(t, 3, il.Query(Eq[uint32]("isnew", false)))
	assert.Equal(t, 1, il.Query(Eq[uint32]("isnew", true)))
}
