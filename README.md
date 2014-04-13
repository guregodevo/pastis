pastis
======

Golang framework for developing ops-friendly, high-performance, RESTful web services


Getting Started
===============

Pastis is a framework for quickly creating RESTful applications with minimal effort: 

```golang
package main

import "net/url"
import "github.com/guregodevo/pastis"

func main() {
	api := pastis.NewAPI()
	api.Do("GET", func(vals url.Values, input int) (int, interface{}) {
		return 200, "Hello"
	}, "/foo")
	api.Start(4567)
}
```

And run with:

```golang
go run main.go
```

View at: http://localhost:4567
