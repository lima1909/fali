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

type ErrInvalidOperation struct {
	indexName string
	op        Op
}

func (e ErrInvalidOperation) Error() string {
	return fmt.Sprintf("index: %q doesn't support the operation: %s", e.indexName, e.op)
}

type ErrValueNotFound struct{ value any }

func (e ErrValueNotFound) Error() string {
	return fmt.Sprintf("index value not found: %v", e.value)
}

type ErrNoIdIndexDefined struct{}

func (e ErrNoIdIndexDefined) Error() string {
	return fmt.Sprintln("no ID index defined")
}

type ErrInvalidArgsLen struct {
	defined string
	got     int
}

func (e ErrInvalidArgsLen) Error() string {
	return fmt.Sprintf("expected: %s values, got: %d", e.defined, e.got)
}
