package uniapi

import (
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
)

// Paginator interface now includes a Merge method for aggregating page results.
// Implementations of Paginator are free to use any strategy (cursor, skip/limit, etc.)
type Paginator interface {
	// Execute should inspect the current aggregated result and request,
	// and return:
	//   newResult: the next page of results (as decoded from JSON)
	//   newRequest: the HTTP request to execute for the next page,
	//   morePages: true if there are additional pages.
	//
	// The base url is provided in case the pagination method requires modifying the url,
	// as is the case with query and path-based pagination.
	//
	// Implementations may modify the options argument.
	Execute(currentResult any, baseUrl string, options *Options, currentRequest *http.Request, middlewares []Middleware) (newRequest *http.Request, morePages bool)

	// Merge should combine the current aggregated result with a new page of results.
	// This is needed because different APIs return different structures.
	Merge(aggregated any, newPage any) (any, int)
}

type SkipLimitPaginator struct {
	// The field name for the list property of the result struct.
	//
	// This is the field that contains the list of items.
	ListFieldName string
	// The field name for the count property of the result struct.
	CountFieldName string
	// The field name for the skip property of the result struct.
	SkipFieldName string
	// The field name for the limit property of the result struct.
	LimitFieldName string

	// The name for the skip query parameter.
	//
	// According to HTTP specs this is case-sensitive, but some APIs may treat them as case-insensitive.
	SkipQueryName  string
	LimitQueryName string
}

func (o *SkipLimitPaginator) Execute(currentResult any, endpointBaseUrl string, options *Options, currentRequest *http.Request, middlewares []Middleware) (newRequest *http.Request, morePages bool) {
	var count, skip, limit int
	var ok bool

	v := reflect.ValueOf(currentResult).Elem()
	countInterface := v.FieldByName(o.CountFieldName).Interface()
	if count, ok = countInterface.(int); !ok {
		log.Printf("unable to cast count field %q to an int, was of type %#v, ending pagination", o.CountFieldName, countInterface)
		return nil, false
	}
	skipInterface := v.FieldByName(o.SkipFieldName).Interface()
	if skip, ok = skipInterface.(int); !ok {
		log.Printf("unable to cast skip field %q to an int, was of type %#v, ending pagination", o.CountFieldName, countInterface)
		return nil, false
	}
	limitInterface := v.FieldByName(o.LimitFieldName).Interface()
	if limit, ok = limitInterface.(int); !ok {
		log.Printf("unable to cast limit field %q to an int, was of type %#v, ending pagination", o.CountFieldName, countInterface)
		return nil, false
	}

	nextSkip := skip + limit
	if nextSkip >= count {
		return nil, false
	}
	log.Printf("skip: %d, limit: %d, count: %d", nextSkip, limit, count)

	if options.QueryParameters == nil {
		options.QueryParameters = make(url.Values)
	}
	options.QueryParameters[o.SkipQueryName] = []string{strconv.Itoa(nextSkip)}
	options.QueryParameters[o.LimitQueryName] = []string{strconv.Itoa(limit)}
	log.Printf("%#v", options.QueryParameters)
	var err error
	url := ApplyOptionsToURL(endpointBaseUrl, options)
	newRequest, err = BuildRequest(currentRequest.Method, url, options, middlewares)
	log.Printf("url: %s", url)
	if err != nil {
		log.Printf("failed building request: %v", err)
		return nil, false
	}
	log.Printf("%#v", newRequest)
	log.Printf("%#v", newRequest.URL.String())

	return newRequest, true
}

func (o *SkipLimitPaginator) Merge(aggregated any, newPage any) (any, int) {
	a := reflect.ValueOf(aggregated).Elem()
	b := reflect.ValueOf(newPage).Elem()
	aList := a.FieldByName(o.ListFieldName)
	bList := b.FieldByName(o.ListFieldName)
	resList := reflect.AppendSlice(aList, bList)
	aList.Set(resList)
	a.FieldByName(o.SkipFieldName).Set(b.FieldByName(o.SkipFieldName))
	return aggregated, resList.Len()
}
