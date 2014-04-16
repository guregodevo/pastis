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
	api.Get( func(vals url.Values) (int, interface{}) {
		return 200, "Hello"
	}, "/foo")
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

	api.Get( func(vals url.Values) (int, interface{}) {
		...show something
	}, "/foo")

	api.Post( func(vals url.Values) (int, interface{}) {
		...create something
	}, "/foo")

	api.Put( func(vals url.Values) (int, interface{}) {
		...modify something
	}, "/foo")

	api.Patch( func(vals url.Values) (int, interface{}) {
		...modify something
	}, "/foo")

	api.Delete( func(vals url.Values) (int, interface{}) {
		...delete something
	}, "/foo")

	api.Link( func(vals url.Values) (int, interface{}) {
		...affiliate something
	}, "/foo")

	api.Unlink( func(vals url.Values) (int, interface{}) {
		...separate something
	}, "/foo")
```

Routes are matched in the order they are defined. The first route that matches the request is invoked.

In Pastis, query parameters or path parameters are both accessible via the first block parameter of type url.Values.

Route patterns may include named parameters:

```go
	api.Get(func(params url.Values) (int, interface{}) {
		title := vals.Get("title")
                ...show something with this named parameter
	}, "/posts/:title")
```

Routes may also utilize query parameters:

```go
	api.get(func(params url.Values) (int, interface{}) {
		title := vals.Get("title")
		author := vals.Get("author")
		greeding := fmt.SPrintf("Hello %s", name)	
		// uses title and author variables; query is optional to the /posts route
	}, "/posts")
```

In Pastis, the request body are also accessible via the second block parameter.
Only data structures that can be represented as valid JSON will be decoded and passed to the second parameters meaning that the type could be : 
 * map[string]interface{}  or of struct types (those that begin with uppercase letter) for JSON Objects
 * []interface{}  for JSON arrays
 * A primitive type int or string in other cases








