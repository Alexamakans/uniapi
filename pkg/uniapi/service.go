package uniapi

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
)

type (
	anyEndpoint           = any
	anyEndpointReturnType = any
)

var (
	ErrNoSuchEndpoint  = errors.New("no endpoint registered")
	ErrInvalidEndpoint = errors.New("invalid endpoint")
)

type Service interface {
	AddEndpoint(method, path string, endpoint anyEndpoint)

	Get(path string, options Options) (anyEndpointReturnType, error)
	Post(path string, options Options) (anyEndpointReturnType, error)
}

func NewService(baseUrl string, middlewares ...Middleware) Service {
	return &service{
		middlewares:   middlewares,
		getEndpoints:  make(map[string]anyEndpoint),
		postEndpoints: make(map[string]anyEndpoint),
	}
}

type service struct {
	middlewares   []Middleware
	getEndpoints  map[string]anyEndpoint
	postEndpoints map[string]anyEndpoint
}

func (o *service) AddEndpoint(method, path string, endpoint anyEndpoint) {
}

func (o *service) Get(path string, options *Options) (any, error) {
	endpoint, ok := o.getEndpoints[path]
	if !ok {
		return nil, fmt.Errorf("GET %s: %w", path, ErrNoSuchEndpoint)
	}
	callField := reflect.ValueOf(endpoint).FieldByName("Call")
	if !callField.IsValid() {
		return nil, fmt.Errorf("GET %s: Call property is invalid, must be a method: %w", path, ErrInvalidEndpoint)
	}

	values := callField.Call([]reflect.Value{reflect.ValueOf(http.MethodGet), reflect.ValueOf(options)})
	result := values[0].Interface()
	err := values[1].Interface().(error)
	if err != nil {
		return nil, err
	}
}

func Call[T any](service Service, method, path string, options Options) (*T, error) {
	var result any
	var err error
	switch method {
	case http.MethodGet:
		result, err = service.Get(path, options)
	case http.MethodPost:
		result, err = service.Post(path, options)
	}

	if err != nil {
		return nil, err
	}

	castResult, ok := result.(*T)
	if !ok {
		return nil, fmt.Errorf("unable to cast result to %q", reflect.TypeOf(*new(T)).Name())
	}

	return castResult, nil
}
