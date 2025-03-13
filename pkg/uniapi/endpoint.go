package uniapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrUnauthenticated = errors.New("401 unauthorized")
	ErrUnauthorized    = errors.New("403 forbidden")
)

type Options struct {
	// PathExtension is used for APIs that take path parameters.
	// Example: {baseUrl}/{baseApiPath}/{threats|endpoints}/{id}
	//                                 ^PathExtension starts here
	PathExtension string
	// Query is the query parameters in the format 'key1=value1&key2=value2' etc.
	Query string
	// Body is the JSON body to send in the request, if any.
	Body []byte

	ExtraHeaders map[string]string
}

type Endpoint[T any] interface {
	Call(method, url string, options *Options) (T, error)
}

type BaseEndpoint[T any] struct{}

// Call constructs and sends an HTTP request using net/http.
// It returns a result of type T which is decoded from the JSON response.
func (o *BaseEndpoint[T]) Call(method, url string, options *Options) (T, error) {
	var result T

	// Build the URL.
	url = fmt.Sprintf("%s%s", url, options.PathExtension)
	if options.Query != "" {
		url = fmt.Sprintf("%s?%s", url, options.Query)
	}

	// Create the HTTP request.
	req, err := http.NewRequest(method, url, bytes.NewBuffer(options.Body))
	if err != nil {
		return result, err
	}

	// Optionally, set content type header if body is provided.
	if len(options.Body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set custom headers if provided.
	for key, value := range options.ExtraHeaders {
		req.Header.Set(key, value)
	}

	// Use a default http.Client.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	// Check for common HTTP error status codes.
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return result, ErrUnauthenticated
	case http.StatusForbidden:
		return result, ErrUnauthorized
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return result, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode JSON response into the generic type T.
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&result); err != nil {
		return result, err
	}

	return result, nil
}
