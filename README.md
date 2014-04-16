pastis
======

Go framework for developing ops-friendly, high-performance, RESTful web services


Getting Started
===============

Pastis is a framework for quickly creating RESTful applications with minimal effort: 

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

Routes
======

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


```

