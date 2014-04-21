pastis
======

Go framework for developing ops-friendly, high-performance, RESTful web services


Getting Started
===============

Pastis is a framework for quickly creating RESTful applications with minimal effort: 

##Quick Example

```go
//main.go
package main

import "net/url"
import "github.com/guregodevo/pastis"

func main() {
	api := pastis.NewAPI()
	api.Get("/foo",  func() (int, interface{}) {
		return 200, "Hello"
	})
	api.Start(4567)
}
```

And run with:

```
go run main.go
```

View at: http://localhost:4567/foo

##Routes

In Pastis, a route is an HTTP method paired with a URL-matching pattern.
Each route is associated with a callback function: 

```go

	api.Get("/foo", func(params url.Values) (int, interface{}) {
		...show something
	})

	api.Post("/foo", func(params url.Values) (int, interface{}) {
		...create something
	})

	api.Put("/foo", func(params url.Values) (int, interface{}) {
		...modify something
	})

	api.Patch("/foo", func(params url.Values) (int, interface{}) {
		...modify something
	})

	api.Delete("/foo", func(params url.Values) (int, interface{}) {
		...delete something
	})

	api.Link("/foo", func(params url.Values) (int, interface{}) {
		...affiliate something
	})

	api.Unlink("/foo", func(params url.Values) (int, interface{}) {
		...separate something
	})
```

Routes are matched in the order they are defined. The first route that matches the request is invoked.

In Pastis, query or path parameters are both accessible via the optional callback parameter of type url.Values. Note that this parameter is optional and there must be at most one of this type among the callback input parameters. By convention, it must be declared before any other callback parameter.

Route patterns may include named parameters:

```go
	api.Get("/posts/:title", func(params url.Values) (int, interface{}) {
		title := params.Get("title")
                ...show something with this named parameter
	})
```

Routes may also utilize query parameters:

```go
	api.Get("/posts", func(params url.Values) (int, interface{}) {
		title := params.Get("title")
		author := params.Get("author")
		greeding := fmt.SPrintf("Hello %s", name)	
		// uses title and author variables; query is optional to the /posts route
	})
```

Routes may require the request body. In Pastis, the request body is decoded to the type of the callback parameter that you declared as input parameter in the callback. Any parameter that has a type different from url.Values will match the request body content provided that it can be represented as valid JSON. 

Possible request body parameter can be any of the following types: 
 * map[string]interface{}  or struct (those that begin with uppercase letter) for JSON Objects
 * []interface{}  for JSON arrays
 * Any Go type that matches the body content that is more convenient that the type above (int, string etc..)

##Return Values

Every callback execution should end up returning a tuple (int, interface{}). The tuple element of type int represents the HTTP status code. The other one of type interface{} represents the response content. The return handler will take care of marshalling this content into JSON.

Examples:
```go
	return http.StatusOK, [] Chart{Chart{"name", 1},Chart{"name", 1}}
	return http.StatusCreated, Identifier{params.Get("id"), 2}
	return http.StatusCreated, map[string]interface{} {"id":1, "size":3, "type":"line"}
	return http.StatusOK, "Hello"
```

##Resources

In Pastis, a resource is any Go struct that implements one of the HTTP method. 

```go
type DashboardResource struct {
}

type ChartResource struct {
}

type Chart struct {
	Name  string
	Order int
}

func (api DashboardResource) GET(params url.Values) (int, interface{}) {
	...do something with params params.Get("dashboardid")	
	return http.StatusOK, [] Chart{Chart{"name", 1},Chart{"name", 1}}
}

func (api ChartResource) GET(params url.Values) (int, interface{}) {
	return http.StatusOK, Chart{params.Get("chartid"), 2}
}

func (api ChartResource) PUT(params url.Values) (int, interface{}) {
	...do something with params params.Get("chartid")
}
```

A resource has a unique URL-matching pattern. Therefore, each resource route method is associated with the resource method function whose name matches.

```go
dashboardResource := new(DashboardResource)
chartResource := new(ChartResource)
api := NewAPI()
api.AddResource("/dashboards/:dashboardid", dashboardResource)
api.AddResource("/dashboards/:dashboardid/charts/:chartid", chartResource )
api.Start(44444)
```

In the above example, the chart resource PUT method matches the HTTP method "PUT" and the resource URL  "/dashboards/:dashboardid/charts/:chartid". 

Resource method functions behave exactly like callback method except that they match the resource route.

##Filters

Filters are evaluated before and/or after request within the same context as the routes will be and can modify the request and response.

A filter is any function that sastifies this interface : 

```go
type Filter func(http.ResponseWriter, *http.Request, *FilterChain)


// Filter (post-process) Filter (as a struct that defines a FilterFunction)
func LoggingFilter(w http.ResponseWriter, request *http.Request, chain *FilterChain) {
	now := time.Now()
	chain.NextFilter(w, request)
	log.Printf("[HTTP] %s %s [%v]\n", request.Method, request.URL, time.Now().Sub(now))
}

```

Any filter can be added to apis

```go
	var api = pastis.NewAPI()
	api.AddFilter(pastis.LoggingFilter)
```

##CORS Support

Pastis provides CORS filter. If you need it, just add the CORS filter to your api. 

```go
	var api = pastis.NewAPI()
	api.AddFilter(pastis.CORSFilter)
```


## Testing

Pastis tests can be written using any testing library or framework. The native Go package httptest is recommended:

```go
import (
	"net/http/httptest"
	"reflect"
	"testing"
)

/* Test Helpers */
func expect(t *testing.T, a interface{}, b interface{}) {
	....
}


func assert_HTTP_Response(t *testing.T, res *http.Response, expectedStatusCode int, expectedResponsebody interface{}) {
	....
}

func Test_Callback_With_Params(t *testing.T) {
	p := NewAPI()
	p.Get( "/hello/:name", func(params url.Values) (int, interface{}) {
		fmt.Printf("Name : %v",params.Get("name"))
		return http.StatusOK, Foo { params.Get("name"), 1 }
	})
	p.HandleFunc()

	ts := httptest.NewServer(p)
	defer ts.Close()

	url := ts.URL + "/hello/guregodevo"
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	assert_HTTP_Response(t, res, http.StatusOK, Foo{"guregodevo", 1})
}
```




