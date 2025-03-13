package uniapi

import (
	"errors"
	"fmt"
	"net/http"
)

type MiddlewareError struct {
	Code int
	Err  error
}

func (o *MiddlewareError) Error() string {
	return fmt.Sprintf("%d: %s", o.Code, o.Err.Error())
}

func (o *MiddlewareError) Unwrap() error {
	return o.Err
}

type Middleware interface {
	Execute(req *http.Request) *http.Request
}

type UnauthenticatedMiddleware struct{}

func (o *UnauthenticatedMiddleware) Execute(req *http.Request) (*http.Request, error) {
	return req, nil
}

var ErrEmptyBearerToken = errors.New("bearer token is empty")

type BearerAuthMiddleware struct {
	bearerPrefixedB64token string
}

func (o *BearerAuthMiddleware) Execute(req *http.Request) (*http.Request, error) {
	if o.bearerPrefixedB64token == "" {
		return req, &MiddlewareError{
			Code: http.StatusUnauthorized,
			Err:  ErrEmptyBearerToken,
		}
	}
	req.Header.Set("Authorization", o.bearerPrefixedB64token)
	return req, nil
}
