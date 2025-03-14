package uniapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var (
	ErrUnauthenticated = errors.New("401 unauthorized")
	ErrUnauthorized    = errors.New("403 forbidden")
)

type Options struct {
	// PathExtension is used for APIs where the parameters are part of the path.
	PathExtension []any
	// Query is the query parameters in the format 'key1=value1&key2=value2' etc.
	QueryParameters url.Values
	// Body is the JSON body to send in the request, if any.
	//
	// The Content-Type header will be set to 'application/json' before
	// any extra headers are applied.
	Body []byte

	ExtraHeaders map[string]string
}

type Endpoint interface {
	GetPath() string

	Call(method, url string, options *Options, middlewares []Middleware) (anyEndpointReturnTypePointer, error)
}

type BaseEndpoint[T any] struct {
	Path      string
	Paginator Paginator
}

func (o *BaseEndpoint[T]) GetPath() string {
	return o.Path
}

// Call constructs and sends an HTTP request using net/http.
// It returns a pointer to a result of type T which is decoded from the JSON response.
// If a paginator is set, it will attempt to retrieve additional pages and merge them.
//
// If the endpoint requires pagination then the options argument may be modified!
func (o *BaseEndpoint[T]) Call(method, baseUrl string, options *Options, middlewares []Middleware) (anyEndpointReturnTypePointer, error) {
	// Create an initial result.
	result := new(T)

	// Build the URL using the endpoint's Path.
	url := fmt.Sprintf("%s/%s", strings.TrimRight(baseUrl, "/"), strings.TrimLeft(o.Path, "/"))
	endpointBaseUrl := url
	url = ApplyOptionsToURL(url, options)

	req, err := BuildRequest(method, url, options, middlewares)
	if err != nil {
		return result, fmt.Errorf("failed building request: %w", err)
	}

	client := &http.Client{}

	// Execute the initial request.
	if err := doRequest(req, client, result); err != nil {
		return result, err
	}

	// If a paginator is provided, attempt to fetch additional pages.
	if o.Paginator != nil {
		aggregated := any(result) // starting aggregated result
		currentReq := req         // using the initial request as context; paginator may override this

		// Loop until there are no more pages.
		for {
			newReq, morePages := o.Paginator.Execute(aggregated, endpointBaseUrl, options, currentReq, middlewares)
			if !morePages {
				break
			}

			// Decode the next page into a temporary interface.
			pageData := new(T)
			if err := doRequest(newReq, client, pageData); err != nil {
				return result, err
			}

			// Merge the new page into the aggregated result.
			var count int
			aggregated, count = o.Paginator.Merge(aggregated, pageData)
			log.Printf("count: %d", count)

			// Set currentReq to newReq for the next iteration.
			currentReq = newReq
		}

		if final, ok := aggregated.(*T); ok {
			result = final
		} else {
			return result, fmt.Errorf("aggregated result is not of expected type")
		}
	}

	return result, nil
}

func ApplyOptionsToURL(url string, options *Options) string {
	for _, extension := range options.PathExtension {
		url = strings.TrimRight(url, "/")
		url = fmt.Sprintf("%s/%v", url, extension)
	}
	url = strings.TrimRight(url, "/")
	if options.QueryParameters != nil {
		url = fmt.Sprintf("%s?%s", url, options.QueryParameters.Encode())
	}
	return url
}

func BuildRequest(method, url string, options *Options, middlewares []Middleware) (*http.Request, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(options.Body))
	if err != nil {
		return nil, err
	}

	if len(options.Body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	for key, value := range options.ExtraHeaders {
		req.Header.Set(key, value)
	}

	for _, middleware := range middlewares {
		req, err = middleware.Execute(req)
		if err != nil {
			return nil, fmt.Errorf("executing middleware: %w", err)
		}
	}

	return req, nil
}

func doRequest(req *http.Request, client *http.Client, result any) error {
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return ErrUnauthenticated
	case http.StatusForbidden:
		return ErrUnauthorized
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	log.Printf("%s", req.URL.String())
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(result); err != nil {
		return err
	}
	return nil
}
