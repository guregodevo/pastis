package pastis

import (
	"log"
	"net/http"
	"time"
)

// FilterChain is a request scoped object to process one or more filters before calling the target function.
type FilterChain struct {
	Filters []Filter         // ordered list of Filter function
	Index   int              // index into filters that is currently in progress
	Target  http.HandlerFunc // function to call after passing all filters
}

func (f *FilterChain) Copy() FilterChain {
	return FilterChain{f.Filters, f.Index, f.Target}
}

// NextFilter passes the request,response pair through the next of Filters.
// Each filter can decide to proceed to the next Filter or handle the Response itself.
func (f *FilterChain) NextFilter(rw http.ResponseWriter, request *http.Request) {
	if f.Index < len(f.Filters) {
		f.Index++
		f.Filters[f.Index-1](rw, request, f)
	} else {
		f.Target(rw, request)
	}
}

// Return the request handler calling the filter after passing through its own filters
func (f *FilterChain) dispatchRequestHandler() http.HandlerFunc {
	return func(rw http.ResponseWriter, request *http.Request) {
		f.Index = 0
		if len(f.Filters) > 0 {
			f.NextFilter(rw, request)
		} else {
			// unfiltered
			f.Target(rw, request)
		}
	}
}

// Filter function definition. Must be called on the FilterChain to pass on the control and eventually call the Target function
type Filter func(http.ResponseWriter, *http.Request, *FilterChain)

// Filter (post-process) Filter (as a struct that defines a FilterFunction)
func LoggingFilter(w http.ResponseWriter, request *http.Request, chain *FilterChain) {
	now := time.Now()
	chain.NextFilter(w, request)
	log.Printf("[HTTP] %s %s [%v]\n", request.Method, request.URL, time.Now().Sub(now))
}
