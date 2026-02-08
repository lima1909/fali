package main

import (
	"fmt"
	"reflect"
)

type ErrInvalidIndexdName struct{ fieldName string }

func (e ErrInvalidIndexdName) Error() string {
	return fmt.Sprintf("could not found index for field name: %s", e.fieldName)
}

type ErrInvalidIndexValue[V any] struct{ value any }

func (e ErrInvalidIndexValue[V]) Error() string {
	return fmt.Sprintf("invalid index value type: %T, expected type: %v", e.value, reflect.TypeFor[V]())
}

type ErrInvalidRelation struct{ relation Relation }

func (e ErrInvalidRelation) Error() string {
	return fmt.Sprintf("this index doesn't support this relation: %#v", e.relation)
}

type ErrValueNotFound struct{ value any }

func (e ErrValueNotFound) Error() string {
	return fmt.Sprintf("index value not found: %v", e.value)
}
