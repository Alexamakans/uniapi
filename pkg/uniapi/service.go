package uniapi

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
)

// Define types for endpoints.
type (
	anyEndpointReturnTypePointer = any
)

var (
	ErrNoSuchEndpoint  = errors.New("no endpoint registered")
	ErrInvalidEndpoint = errors.New("invalid endpoint")
)

// Service defines the API service interface.
type Service interface {
	AddEndpoint(method string, endpoint Endpoint)
	Get(path string, options *Options) (anyEndpointReturnTypePointer, error)
	Post(path string, options *Options) (anyEndpointReturnTypePointer, error)
}

// NewService returns a new service instance.
func NewService(baseUrl string, middlewares ...Middleware) Service {
	return &service{
		baseUrl:       baseUrl,
		middlewares:   middlewares,
		getEndpoints:  make(map[string]Endpoint),
		postEndpoints: make(map[string]Endpoint),
	}
}

type service struct {
	baseUrl       string
	middlewares   []Middleware
	getEndpoints  map[string]Endpoint
	postEndpoints map[string]Endpoint
}

// AddEndpoint registers an endpoint for a specific HTTP method and path.
func (o *service) AddEndpoint(method string, endpoint Endpoint) {
	switch method {
	case http.MethodGet:
		o.getEndpoints[endpoint.GetPath()] = endpoint
	case http.MethodPost:
		o.postEndpoints[endpoint.GetPath()] = endpoint
	default:
		// You might choose to support other methods or log an error.
	}
}

// Get looks up and calls the GET endpoint for the given path.
func (o *service) Get(path string, options *Options) (anyEndpointReturnTypePointer, error) {
	endpoint, ok := o.getEndpoints[path]
	if !ok {
		return nil, fmt.Errorf("GET %s: %w", path, ErrNoSuchEndpoint)
	}

	// Retrieve the "Call" method by reflection.
	callMethod := reflect.ValueOf(endpoint).MethodByName("Call")
	if !callMethod.IsValid() {
		return nil, fmt.Errorf("GET %s: Call method is invalid, must be a method: %w", path, ErrInvalidEndpoint)
	}

	// Call the method with two arguments: HTTP method and options.
	argValues := []reflect.Value{
		reflect.ValueOf(http.MethodGet),
		reflect.ValueOf(o.baseUrl),
		reflect.ValueOf(options),
		reflect.ValueOf(o.middlewares),
	}
	returnValues := callMethod.Call(argValues)
	result := returnValues[0].Interface()
	var err error
	if !returnValues[1].IsNil() {
		err = returnValues[1].Interface().(error)
	}
	return result, err
}

// Post looks up and calls the POST endpoint for the given path.
func (o *service) Post(path string, options *Options) (anyEndpointReturnTypePointer, error) {
	endpoint, ok := o.postEndpoints[path]
	if !ok {
		return nil, fmt.Errorf("POST %s: %w", path, ErrNoSuchEndpoint)
	}

	// Retrieve the "Call" method by reflection.
	callMethod := reflect.ValueOf(endpoint).MethodByName("Call")
	if !callMethod.IsValid() {
		return nil, fmt.Errorf("POST %s: Call method is invalid, must be a method: %w", path, ErrInvalidEndpoint)
	}

	// Call the method with two arguments: HTTP method and options.
	returnValues := callMethod.Call([]reflect.Value{
		reflect.ValueOf(http.MethodPost),
		reflect.ValueOf(options),
	})
	result := returnValues[0].Interface()
	var err error
	if !returnValues[1].IsNil() {
		err = returnValues[1].Interface().(error)
	}
	return result, err
}

// Call is a generic helper function to call a service endpoint
// and cast its result to the expected type T.
func Call[T any](service Service, method, path string, options Options) (*T, error) {
	var result any
	var err error
	// Pass a pointer to options.
	switch method {
	case http.MethodGet:
		result, err = service.Get(path, &options)
	case http.MethodPost:
		result, err = service.Post(path, &options)
	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
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
