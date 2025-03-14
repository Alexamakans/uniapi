package uniapi

import (
	"errors"
	"fmt"
	"net/http"
)

type Middleware interface {
	Execute(req *http.Request) (*http.Request, error)
}

type UnauthenticatedMiddleware struct{}

func (o UnauthenticatedMiddleware) Execute(req *http.Request) (*http.Request, error) {
	return req, nil
}

var ErrEmptyBearerToken = errors.New("bearer token is empty")

type BearerAuthMiddleware struct {
	bearerPrefixedB64token string
}

func (o BearerAuthMiddleware) Execute(req *http.Request) (*http.Request, error) {
	if o.bearerPrefixedB64token == "" {
		return req, fmt.Errorf("%w: %w", ErrUnauthenticated, ErrEmptyBearerToken)
	}
	req.Header.Set("Authorization", o.bearerPrefixedB64token)
	return req, nil
}
