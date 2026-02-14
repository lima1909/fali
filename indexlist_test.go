package main

import (
	"strings"
	"sync"
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

	il.CreateIndex("name", NewMapIndex(func(c *car) string { return c.name }))
	il.CreateIndex("isnew", NewMapIndex(func(c *car) bool { return c.isNew }))

	il.Add(car{name: "Dacia", age: 22, color: "red"})
	il.Add(car{name: "Opel", age: 22})
	il.Add(car{name: "Mercedes", age: 5, isNew: true})
	il.Add(car{name: "Dacia", age: 22})

	il.CreateIndex("age", NewMapIndex(func(c *car) uint8 { return c.age }))

	c, found := il.list.Get(1)
	assert.True(t, found)
	assert.Equal(t, car{name: "Opel", age: 22}, c)

	_, found = il.list.Get(99)
	assert.False(t, found)

	qr, err := il.Query(Eq("name", "Opel"))
	assert.NoError(t, err)
	assert.Equal(t, 1, qr.Count())

	qr, err = il.Query(Eq("age", uint8(5)))
	assert.NoError(t, err)
	assert.Equal(t, 1, qr.Count())

	qr, err = il.Query(Eq("isnew", false))
	assert.NoError(t, err)
	assert.Equal(t, 3, qr.Count())

	qr, err = il.Query(Eq("isnew", true))
	assert.NoError(t, err)
	assert.Equal(t, 1, qr.Count())
	// wrong value type, expected: uint8, got int

	qr, err = il.Query(Eq("age", 5))
	assert.Error(t, err)
	assert.Equal(t, QueryResult[car]{}, qr)

	// wrong field name, expected: age, got wrong
	qr, err = il.Query(Eq("wrong", 5))
	assert.Error(t, err)
	assert.Equal(t, QueryResult[car]{}, qr)
}

func TestIndexList_QueryResult(t *testing.T) {
	less := func(c1, c2 *car) bool { return strings.Compare(c1.name, c2.name) < 0 }

	il := NewIndexList[car]()
	il.CreateIndex("age", NewMapIndex(func(c *car) uint8 { return c.age }))

	il.Add(car{name: "Mercedes", age: 22, color: "red"})
	il.Add(car{name: "Opel", age: 22})
	il.Add(car{name: "Dacia", age: 5, isNew: true})
	il.Add(car{name: "Dacia", age: 22})
	il.Add(car{name: "Audi", age: 22})

	qr, err := il.Query(Eq("age", uint8(22)))
	assert.NoError(t, err)

	assert.False(t, qr.Empty())
	assert.Equal(t, 4, qr.Count())

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
	il.CreateIndex("name", NewMapIndex(func(c *car) string { return c.name }))
	il.CreateIndex("age", NewMapIndex(func(c *car) uint8 { return c.age }))

	il.Add(car{name: "Mercedes", age: 22, color: "red"})
	il.Add(car{name: "Opel", age: 22})
	il.Add(car{name: "Dacia", age: 5, isNew: true})
	il.Add(car{name: "Dacia", age: 22})
	il.Add(car{name: "Audi", age: 22})

	qr, err := il.Query(All())
	assert.NoError(t, err)

	assert.False(t, qr.Empty())
	assert.Equal(t, 5, qr.Count())

	// remove item on index 3
	c, removed := il.removeNoLock(3)
	assert.True(t, removed)
	assert.Equal(t, 4, il.Count())
	assert.Equal(t, c, car{name: "Dacia", age: 22})

	// try to find item on index 3
	qr, err = il.Query(And(Eq("name", "Dacia"), Eq("age", uint8(22))))
	assert.NoError(t, err)
	assert.Equal(t, 0, qr.Count())

	_, removed = il.removeNoLock(99)
	assert.False(t, removed)

	qr, err = il.Query(Eq("name", "Dacia"))
	assert.NoError(t, err)
	assert.Equal(t, 1, qr.Count())
	assert.Equal(t, []car{{name: "Dacia", age: 5, isNew: true}}, qr.Values())

	qr, err = il.Query(Eq("age", uint8(22)))
	assert.NoError(t, err)
	assert.Equal(t, 3, qr.Count())

	qr.Remove()
	assert.Equal(t, 1, il.Count())

	c, found := il.list.Get(2)
	assert.True(t, found)
	assert.Equal(t, car{name: "Dacia", age: 5, isNew: true}, c)
}

func TestIndexList_RemoveLater(t *testing.T) {
	il := NewIndexList[car]()
	il.CreateIndex("name", NewMapIndex(func(c *car) string { return c.name }))
	il.CreateIndex("age", NewMapIndex(func(c *car) uint8 { return c.age }))

	il.Add(car{name: "Mercedes", age: 22, color: "red"})
	il.Add(car{name: "Opel", age: 22})
	il.Add(car{name: "Dacia", age: 5, isNew: true})
	il.Add(car{name: "Dacia", age: 22})
	il.Add(car{name: "Audi", age: 22})

	qr1, err := il.Query(Eq("name", "Dacia"))
	assert.NoError(t, err)
	assert.Equal(t, 2, qr1.Count())

	qr2, err := il.Query(Eq("name", "Dacia"))
	assert.NoError(t, err)
	assert.Equal(t, 2, qr2.Count())

	qr1.Remove()
	assert.Equal(t, 0, qr1.Count())
	assert.Equal(t, 0, qr2.Count())

	_, err = il.Query(Eq("name", "Dacia"))
	// Dacia doesn't exist anymore
	assert.ErrorIs(t, ErrValueNotFound{"Dacia"}, err)

	// qr1 has allready remove all Dacia
	qr2.Remove()
}

func TestIndexList_RemoveLaterAsync(t *testing.T) {
	il := NewIndexList[car]()
	il.CreateIndex("name", NewMapIndex(func(c *car) string { return c.name }))
	il.CreateIndex("age", NewMapIndex(func(c *car) uint8 { return c.age }))

	il.Add(car{name: "Mercedes", age: 22, color: "red"})
	il.Add(car{name: "Opel", age: 22})
	il.Add(car{name: "Dacia", age: 5, isNew: true})
	il.Add(car{name: "Dacia", age: 22})
	il.Add(car{name: "Audi", age: 22})

	qr1, err := il.Query(Eq("name", "Dacia"))
	assert.NoError(t, err)
	assert.Equal(t, 2, qr1.Count())

	qr2, err := il.Query(Eq("name", "Dacia"))
	assert.NoError(t, err)
	assert.Equal(t, 2, qr2.Count())

	var wg sync.WaitGroup

	wg.Go(func() {
		qr1.Remove()
		assert.Equal(t, 0, qr1.Count())
	})

	wg.Go(func() {
		qr2.Remove()
		assert.Equal(t, 0, qr2.Count())
	})

	wg.Wait()
}

func TestIndexList_CreateIndex(t *testing.T) {
	il := NewIndexList[car]()
	il.Add(car{name: "Dacia", age: 22, color: "red"})
	il.Add(car{name: "Opel", age: 22})
	il.Add(car{name: "Mercedes", age: 5, isNew: true})
	il.Add(car{name: "Dacia", age: 22})

	_, err := il.Query(Eq("name", "Opel"))
	assert.Error(t, err)
	assert.Equal(t, "could not found index for field name: name", err.Error())

	// create Index for name
	il.CreateIndex("name", NewMapIndex(func(c *car) string { return c.name }))
	qr, err := il.Query(Eq("name", "Opel"))
	assert.NoError(t, err)
	assert.Equal(t, 1, qr.Count())
	assert.Equal(t, []car{{name: "Opel", age: 22}}, qr.Values())
}

func TestIndexList_CreateIndexVarious(t *testing.T) {
	il := NewIndexList[car]()
	il.CreateIndex("name", NewMapIndex(func(c *car) string { return c.name }))
	il.CreateIndex("age", NewSortedIndex(func(c *car) uint8 { return c.age }))

	il.Add(car{name: "Dacia", age: 2, color: "red"})
	il.Add(car{name: "Opel", age: 12})
	il.Add(car{name: "Mercedes", age: 5, isNew: true})
	il.Add(car{name: "Dacia", age: 22})

	qr, err := il.Query(Eq("name", "Opel"))
	assert.NoError(t, err)
	assert.Equal(t, 1, qr.Count())
	assert.Equal(t, []car{{name: "Opel", age: 12}}, qr.Values())

	qr, err = il.Query(Lt("age", uint8(13)))
	assert.NoError(t, err)
	assert.Equal(t, 3, qr.Count())
	assert.Equal(t, []car{
		{name: "Dacia", age: 2, color: "red"},
		{name: "Opel", age: 12},
		{name: "Mercedes", age: 5, isNew: true},
	}, qr.Values())

	qr, err = il.Query(Le("age", uint8(12)))
	assert.NoError(t, err)
	assert.Equal(t, 3, qr.Count())
	assert.Equal(t, []car{
		{name: "Dacia", age: 2, color: "red"},
		{name: "Opel", age: 12},
		{name: "Mercedes", age: 5, isNew: true},
	}, qr.Values())

	qr, err = il.Query(Gt("age", uint8(11)))
	assert.NoError(t, err)
	assert.Equal(t, 2, qr.Count())
	assert.Equal(t, []car{
		{name: "Opel", age: 12},
		{name: "Dacia", age: 22},
	}, qr.Values())

	qr, err = il.Query(Ge("age", uint8(12)))
	assert.NoError(t, err)
	assert.Equal(t, 2, qr.Count())
	assert.Equal(t, []car{
		{name: "Opel", age: 12},
		{name: "Dacia", age: 22},
	}, qr.Values())
}

func TestIndexList_StringItem(t *testing.T) {
	il := NewIndexList[string]()
	il.CreateIndex("val", NewMapIndex(func(s *string) string { return *s }))

	il.Add("Dacia")
	il.Add("Opel")
	il.Add("Mercedes")
	il.Add("Dacia")

	qr, err := il.Query(Eq("val", "Dacia"))
	assert.NoError(t, err)
	assert.Equal(t, 2, qr.Count())
	assert.Equal(t, []string{"Dacia", "Dacia"}, qr.Values())
}

func TestIndexList_StringPtrItemWithNil(t *testing.T) {
	il := NewIndexList[*string]()
	il.CreateIndex("val", NewMapIndex(func(s **string) *string { return *s }))

	dacia := "Dacia"
	il.Add(&dacia)
	il.Add(nil)
	il.Add(&dacia)

	qr, err := il.Query(Eq("val", &dacia))
	assert.NoError(t, err)
	assert.Equal(t, 2, qr.Count())
	assert.Equal(t, []*string{&dacia, &dacia}, qr.Values())

	// Eq = nil
	qr, err = il.Query(Eq("val", (*string)(nil)))
	assert.NoError(t, err)
	assert.Equal(t, 1, qr.Count())
	assert.Equal(t, []*string{nil}, qr.Values())

	// IsNil
	qr, err = il.Query(IsNil[string]("val"))
	assert.NoError(t, err)
	assert.Equal(t, 1, qr.Count())
	assert.Equal(t, []*string{nil}, qr.Values())

	// Or(IsNil, Eq(dacia)
	qr, err = il.Query(Or(IsNil[string]("val"), Eq("val", &dacia)))
	assert.NoError(t, err)
	assert.Equal(t, 3, qr.Count())
	assert.Equal(t, []*string{&dacia, nil, &dacia}, qr.Values())

	// wrong IsNil Query
	_, err = il.Query(IsNil[*string]("val"))
	assert.ErrorIs(t, err, ErrInvalidIndexValue[*string]{(**string)(nil)})

	_, err = il.Query(IsNil[int]("val"))
	assert.ErrorIs(t, err, ErrInvalidIndexValue[*string]{(*int)(nil)})
}

func TestIndexList_WithID(t *testing.T) {
	il := NewIndexListWithID(func(c *car) string { return c.name })
	il.CreateIndex("isnew", NewMapIndex(func(c *car) bool { return c.isNew }))

	il.Add(car{name: "Opel", age: 22})
	il.Add(car{name: "Mercedes", age: 5, isNew: true})
	il.Add(car{name: "Dacia", age: 42})

	dacia, err := il.Get("Dacia")
	assert.NoError(t, err)
	assert.Equal(t, car{name: "Dacia", age: 42}, dacia)
	assert.Equal(t, 3, il.Count())
	assert.True(t, il.ContainsID("Dacia"))
	assert.False(t, il.ContainsID("NotFound"))

	// remove dacia
	removed, err := il.Remove("Dacia")
	assert.NoError(t, err)
	assert.True(t, removed)
	assert.Equal(t, 2, il.Count())

	// check not found after remove
	_, err = il.Get("Dacia")
	assert.ErrorIs(t, err, ErrValueNotFound{"Dacia"})
	_, err = il.Remove("Dacia")
	assert.ErrorIs(t, err, ErrValueNotFound{"Dacia"})

	// wrong datatype
	_, err = il.Get(5)
	assert.ErrorIs(t, err, ErrInvalidIndexValue[string]{5})
	_, err = il.Remove(5)
	assert.ErrorIs(t, err, ErrInvalidIndexValue[string]{5})
}

func TestIndexList_WithID_Errors(t *testing.T) {
	il := NewIndexList[car]()
	il.Add(car{name: "Dacia", age: 42})

	// Get
	_, err := il.Get("Dacia")
	assert.ErrorIs(t, err, ErrNoIdIndexDefined{})

	// Remove
	_, err = il.Remove("Dacia")
	assert.ErrorIs(t, err, ErrNoIdIndexDefined{})
}
