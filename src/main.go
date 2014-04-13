package main

import "net/url"
//import "github.com/guregodevo/pastis"

func main() {
	api := pastis.NewAPI()
	api.Do("GET", func(vals url.Values, input int) (int, interface{}) {
		return 200, "Hello"
	}, "/foo")
	api.Start(4567)
}