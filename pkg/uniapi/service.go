package uniapi

import (
	"errors"
	"net/http"
)

type Options struct {
	Query string
	Body  []byte

	ExtraHeaders map[string]string
}

var (
	ErrNoSuchEndpoint = errors.New("no endpoint registered")
	ErrNoSuchMethod   = errors.New("no endpoint registered")
)

type Service interface {
	AddEndpoint(method, path string, endpoint any)

	Get(path string, options Options) (any, error)
}

func NewService(baseUrl string, middlewares ...Middleware) Service {
	return &service{
		middlewares: middlewares,
	}
}

type service struct {
	middlewares []Middleware
	endpoints   map[string]map[string]any
}

func (o *service) AddEndpoint(method, path string, endpoint any) {
}

func (o *service) Get(path string) (any, error) {
	methodGroup, ok := o.endpoints[http.MethodGet]
	if !ok {
		return nil, ErrNoSuchEndpoint
	}
}
