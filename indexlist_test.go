package main

import (
	"strings"
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
	il.fieldIndexMap = FieldIndexMap[car, uint32]{
		"name":  {index: NewMapIndex[uint32](), fieldFn: func(c *car) any { return c.name }},
		"age":   {index: NewMapIndex[uint32](), fieldFn: func(c *car) any { return c.age }},
		"isnew": {index: NewMapIndex[uint32](), fieldFn: func(c *car) any { return c.isNew }},
	}

	il.Add(car{name: "Dacia", age: 22, color: "red"})
	il.Add(car{name: "Opel", age: 22})
	il.Add(car{name: "Mercedes", age: 5, isNew: true})
	il.Add(car{name: "Dacia", age: 22})

	c, found := il.Get(1)
	assert.True(t, found)
	assert.Equal(t, car{name: "Opel", age: 22}, c)

	_, found = il.Get(99)
	assert.False(t, found)

	qr, err := il.Query(Eq[uint32]("name", "Opel"))
	assert.NoError(t, err)
	assert.Equal(t, 1, qr.Count())
	assert.Equal(t, []uint32{1}, qr.Indices())

	qr, err = il.Query(Eq[uint32]("age", uint8(5)))
	assert.NoError(t, err)
	assert.Equal(t, 1, qr.Count())
	assert.Equal(t, []uint32{2}, qr.Indices())

	qr, err = il.Query(Eq[uint32]("isnew", false))
	assert.NoError(t, err)
	assert.Equal(t, 3, qr.Count())
	assert.Equal(t, []uint32{0, 1, 3}, qr.Indices())

	qr, err = il.Query(Eq[uint32]("isnew", true))
	assert.NoError(t, err)
	assert.Equal(t, 1, qr.Count())
	assert.Equal(t, []uint32{2}, qr.Indices())

	// wrong value type, expected: uint8, got int
	qr, err = il.Query(Eq[uint32]("age", 5))
	assert.Error(t, err)
	assert.Equal(t, QueryResult[car]{}, qr)

	// wrong field name, expected: age, got wrong
	qr, err = il.Query(Eq[uint32]("wrong", 5))
	assert.Error(t, err)
	assert.Equal(t, QueryResult[car]{}, qr)
}

func TestIndexList_QueryResult(t *testing.T) {
	less := func(c1, c2 *car) bool { return strings.Compare(c1.name, c2.name) < 0 }

	il := NewIndexList[car]()
	il.fieldIndexMap = FieldIndexMap[car, uint32]{
		"age": {index: NewMapIndex[uint32](), fieldFn: func(c *car) any { return c.age }},
	}

	il.Add(car{name: "Mercedes", age: 22, color: "red"})
	il.Add(car{name: "Opel", age: 22})
	il.Add(car{name: "Dacia", age: 5, isNew: true})
	il.Add(car{name: "Dacia", age: 22})
	il.Add(car{name: "Audi", age: 22})

	qr, err := il.Query(Eq[uint32]("age", uint8(22)))
	assert.NoError(t, err)

	assert.False(t, qr.Empty())
	assert.Equal(t, 4, qr.Count())
	assert.Equal(t, []uint32{0, 1, 3, 4}, qr.Indices())

	assert.Equal(t, []car{
		{name: "Mercedes", age: 22, color: "red"},
		{name: "Opel", age: 22},
		{name: "Dacia", age: 22},
		{name: "Audi", age: 22},
	},
		qr.Values(),
	)

	assert.Equal(t, []car{
		{name: "Audi", age: 22},
		{name: "Dacia", age: 22},
		{name: "Mercedes", age: 22, color: "red"},
		{name: "Opel", age: 22},
	},
		qr.Sort(less),
	)
}

func TestIndexList_Remove(t *testing.T) {
	il := NewIndexList[car]()
	il.fieldIndexMap = FieldIndexMap[car, uint32]{
		"name": {index: NewMapIndex[uint32](), fieldFn: func(c *car) any { return c.name }},
		"age":  {index: NewMapIndex[uint32](), fieldFn: func(c *car) any { return c.age }},
	}

	il.Add(car{name: "Mercedes", age: 22, color: "red"})
	il.Add(car{name: "Opel", age: 22})
	il.Add(car{name: "Dacia", age: 5, isNew: true})
	il.Add(car{name: "Dacia", age: 22})
	il.Add(car{name: "Audi", age: 22})

	qr, err := il.Query(All[uint32]())
	assert.NoError(t, err)

	assert.False(t, qr.Empty())
	assert.Equal(t, 5, qr.Count())

	// remove item on index 3
	c, removed := il.Remove(3)
	assert.True(t, removed)
	assert.Equal(t, 4, il.Count())
	assert.Equal(t, c, car{name: "Dacia", age: 22})
	// try to find item on index 3
	qr, err = il.Query(Eq[uint32]("name", "Dacia").And(Eq[uint32]("age", uint8(22))))
	assert.NoError(t, err)
	assert.Equal(t, 0, qr.Count())

	_, removed = il.Remove(99)
	assert.False(t, removed)

	qr, err = il.Query(Eq[uint32]("name", "Dacia"))
	assert.NoError(t, err)
	assert.Equal(t, 1, qr.Count())
	assert.Equal(t, []car{{name: "Dacia", age: 5, isNew: true}}, qr.Values())

	qr, err = il.Query(Eq[uint32]("age", uint8(22)))
	assert.NoError(t, err)
	assert.Equal(t, 3, qr.Count())

	assert.Equal(t, 3, qr.Remove())
	assert.Equal(t, 1, il.Count())

	c, found := il.Get(2)
	assert.True(t, found)
	assert.Equal(t, car{name: "Dacia", age: 5, isNew: true}, c)
}
