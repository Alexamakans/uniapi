// Package uniapi provides a unified interface for registering and executing API endpoints.
package uniapi

import (
	"encoding/json"
	"errors"
)

// HandlerFunc is a function type that processes API calls.
// It accepts a map of parameters and returns either any data type or an error.
type HandlerFunc func(params map[string]any) (any, error)

// endpoints stores registered endpoint handlers by their unique name.
var endpoints = make(map[string]HandlerFunc)

// RegisterEndpoint registers a new API endpoint with a given name and handler.
// It returns an error if the name is empty, the handler is nil, or the endpoint already exists.
func RegisterEndpoint(name string, handler HandlerFunc) error {
	if name == "" {
		return errors.New("endpoint name cannot be empty")
	}
	if handler == nil {
		return errors.New("handler cannot be nil")
	}
	if _, exists := endpoints[name]; exists {
		return errors.New("endpoint already registered")
	}
	endpoints[name] = handler
	return nil
}

// CallEndpoint executes the registered endpoint with the provided parameters.
// The returned data is always marshaled into json.RawMessage to allow for flexible consumption.
func CallEndpoint(name string, params map[string]any) (json.RawMessage, error) {
	handler, exists := endpoints[name]
	if !exists {
		return nil, errors.New("endpoint not found")
	}

	result, err := handler(params)
	if err != nil {
		return nil, err
	}

	// If the result is already a json.RawMessage, return it directly.
	if raw, ok := result.(json.RawMessage); ok {
		return raw, nil
	}

	// Otherwise, marshal the result into JSON.
	b, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}

// CallEndpointUnmarshal calls CallEndpoint and unmarshals the result into a new instance of type T.
func CallEndpointUnmarshal[T any](name string, params map[string]any) (T, error) {
	t := new(T)
	data, err := CallEndpoint(name, params)
	if err != nil {
		return *t, err
	}
	if err := json.Unmarshal(data, t); err != nil {
		return *t, err
	}
	return *t, nil
}

// Example usage:
//
//   type CustomResponse struct {
//       Message string `json:"message"`
//       Code    int    `json:"code"`
//   }
//
//   func myHandler(params map[string]any) (any, error) {
//       // Process the request and return a custom response.
//       return CustomResponse{Message: "Hello, World!", Code: 200}, nil
//   }
//
//   func main() {
//       // Register the endpoint.
//       err := RegisterEndpoint("hello", myHandler)
//       if err != nil {
//           log.Fatal(err)
//       }
//
//       // Execute the endpoint.
//       response, err := CallEndpoint("hello", map[string]any{"name": "Alice"})
//       if err != nil {
//           log.Fatal(err)
//       }
//       fmt.Println(string(response))
//   }
