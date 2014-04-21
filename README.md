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

##Return Handler

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

A resource route is a method paired with its URL-matching pattern. 
Given that a resource might have several methods, each method route is associated with the resource function whose name matches the method. 

```go
dashboardResource := new(DashboardResource)
chartResource := new(ChartResource)
api := NewAPI()
api.AddResource("/dashboards/:dashboardid", dashboardResource)
api.AddResource("/dashboards/:dashboardid/charts/:chartid", chartResource, )
api.Start(44444)
```




