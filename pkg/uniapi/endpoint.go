package uniapi

import (
	"errors"
)

var (
	ErrUnauthenticated = errors.New("401 unauthorized")
	ErrUnauthorized    = errors.New("403 forbidden")
)

type Endpoint[T any] interface {
	Call() (T, error)
	getValueInstance() *T
}

type BaseEndpoint[T any] struct{}

func (o BaseEndpoint[T]) getValueInstance() *T {
	return new(T)
}
