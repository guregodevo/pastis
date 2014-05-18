//Package pastis implements a simple library for building RESTful APIs.
//This package provides you with everything you will need for most of your applications.
//Pastis has 3 building blocks : 
// 1) An API associated to a set of Resource and callback functions. It is paired with an arbitrary port.
// 2) Resource paired with an URL-pattern. Its represents a REST resource.
// 3) A callback paired with a URL-pattern and a request method. 
// Note that a pastis server can support more than one API.
// Pastis rich features are : 
// Nice URL-pattern matching, 
// Path parameter parsing,
// Configurable loggers,
// CORS support,
// Speaks JSON and
// Type-safe request
package pastis

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

// An API manages a group of resources by routing to requests
// to the correct method and URL.
//
// You can instantiate multiple APIs on separate ports. Each API
// will manage its own set of resources.
type API struct {
	//An HTTP mutex to add handlers
	mux    *http.ServeMux
	//A filter chain
	chain  *FilterChain
	//A router
	router *Router
	//A configurable logger
	logger *Logger
}

// NewAPI allocates and returns a new API.
func NewAPI() *API {
	return &API{chain: &FilterChain{[]Filter{}, 0, nil}, mux: http.NewServeMux(), router: NewRouter(), logger: GetLogger("DEBUG")}
}

//SetOutput sets the output destination for the standard logger of the given level.
//Example: api.SetOuput("ERROR", os.StdErr, log.Lmicroseconds)
func (api *API) SetOuput(level string, w io.Writer, flag int) *Logger {
	api.logger.SetOuput(level, w, flag)
	return api.logger
}

//A Pretty Error response
func ErrorResponse(err error) interface{} {
	return map[string]string{"error": err.Error()}
}

//Return an instance of http.HandlerFunc built from  a pair of request method and a callback value.
//The first callback input parameter is the set of URL query and path parameters.
//The second callback input parameter is the unmarshalled JSON body recieved from the request (if it exists).
func (api *API) handleMethodCall(urlValues url.Values, request *http.Request, methodRef reflect.Value) (int, interface{}) {
	api.logger.Debugf("handleMethodCall %s", request.Method)

	if methodRef.Kind() == reflect.Invalid {
		return http.StatusNotImplemented, nil
	}

	methodType := methodRef.Type()
	methodArgSize := methodRef.Type().NumIn()

	api.logger.Debugf("method has %d argument.", methodArgSize)

	if methodArgSize >= 3 {
		api.logger.Errorf("method %v cannot have more than 2 arguments", methodRef)
		return http.StatusNotImplemented, nil
	}

	if methodArgSize == 0 {
		api.logger.Errorf("method %v has no argument. Skip marshalling...", methodRef)
		return api.handleReturn(methodRef, []reflect.Value{})
	}

	valueOfUrlValues := reflect.ValueOf(urlValues)
	methodParameterValues := []reflect.Value{valueOfUrlValues}

	expectedJSONType := methodType.In(0)

	if methodArgSize == 1 {
		if expectedJSONType == valueOfUrlValues.Type() {
			api.logger.Debugf(" method has one argument of type url.Values. Skip marshalling...")
			return api.handleReturn(methodRef, methodParameterValues)
		} else {
			api.logger.Debugf(" method %v has one argument of request body type. ", methodRef)
			methodParameterValues = []reflect.Value{} // will add later the json body as parameter
		}
	} else if methodArgSize == 2 {
		api.logger.Debug(" method first argument is not the request body type.\n")
		expectedJSONType = methodType.In(1)
		api.logger.Debugf(" method second argument is the request body type %v.\n", expectedJSONType)
	}

	expectedJSONValue := reflect.New(expectedJSONType)

	jsonInterface := expectedJSONValue.Interface()

	dec := json.NewDecoder(request.Body)
	for {
		if err := dec.Decode(jsonInterface); err == io.EOF {
			break
		} else if err != nil {
			api.logger.Error(" unable to decode json blob. Check whether parameter type matches json type. \n")
			return http.StatusNotImplemented, nil
		}
	}

	jsonValue := reflect.ValueOf(jsonInterface)
	jsonValueType := jsonValue.Elem().Type()

	if expectedJSONType.Kind() != jsonValue.Elem().Type().Kind() {
		api.logger.Errorf(" Unexpected JSON format. Should be of type '", expectedJSONType.Kind(), "' instead of ", jsonValueType.Kind())
		return http.StatusNotImplemented, nil
	} else if expectedJSONType != jsonValue.Type() {
		methodParameterValues = append(methodParameterValues, jsonValue.Elem())
	} else {
		api.logger.Errorf(" Parameter type mismatches json type. Expected JSON format. ", expectedJSONType)
		return http.StatusNotImplemented, nil
	}
	return api.handleReturn(methodRef, methodParameterValues)
}

//Handles the return values. It converts the array of Value into a tuple (int, interface {} )
func (api *API) handleReturn(methodRef reflect.Value, methodParameterValues []reflect.Value) (int, interface{}) {
	responseValues := methodRef.Call(methodParameterValues)
	if len(responseValues) != 2 {
		api.logger.Errorf(" method %v does not return expected response (int, interface{}).", methodRef)
		return http.StatusNotImplemented, nil
	}
	return int(responseValues[0].Int()), responseValues[1].Interface()
}

//Return an instance of http.HandlerFunc built from  a request method, a URL-pattern matching and a callback function fn.
//The first callback input parameter is the set of URL query and path parameters.
//The second callback input parameter is the unmarshalled JSON body recieved from the request (if it exists).
func (api *API) methodHandler(pattern string, requestMethod string, fn reflect.Value) http.HandlerFunc {
	return func(rw http.ResponseWriter, request *http.Request) {

		code, data := api.handleMethodCall(request.Form, request, fn)

		api.handlerFuncReturn(code, data, rw)
	}
}

//Utility method writing status code and data to the given response
func (api *API) handlerFuncReturn(code int, data interface{}, rw http.ResponseWriter) {
	api.logger.Debugf(" handlerFuncReturn %v", code)

	content, err := json.Marshal(data)
	if err != nil {
		api.logger.Errorf(" handlerFuncReturn could not marshall content [%v]", data)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	//FIXME should be configurable
	rw.Header().Set("Content-Type", "application/json")

	rw.WriteHeader(code)

	if content != nil {
		rw.Write(content)
	}
}

// AddFilter adds a new filter to an API. The API will execute the filter
// before calling the target function.
func (api *API) AddFilter(filter Filter) {
	api.chain.Filters = append(api.chain.Filters, filter)
}

// AddResource adds a new resource to an API. The API will route
// requests that match the given path to its HTTP
// method on the resource.
func (api *API) AddResource(pattern string, resource interface{}) {
	methods := []string{"GET", "Get", "Put", "PUT", "Post", "POST", "Patch", "PATCH", "DELETE", "Delete", "Options", "OPTIONS"}
	for _, requestMethod := range methods {
		methodRef := reflect.ValueOf(resource).MethodByName(requestMethod)
		if methodRef.Kind() != reflect.Invalid {
			requestMethod = strings.ToUpper(requestMethod)
			handler := api.methodHandler(pattern, requestMethod, methodRef)
			api.addHandler(requestMethod, handler, pattern)
			api.logger.Debugf(" Added Resource [method={%v},pattern={%v}]", requestMethod, pattern)
		}
	}
}

// Function callback paired with a request Method and URL-matching pattern.
func (api *API) Do(requestMethod string, pattern string, fn interface{}) {
	handler := api.methodHandler(pattern, requestMethod, reflect.ValueOf(fn))
	api.addHandler(requestMethod, handler, pattern)
	api.logger.Debugf(" Added Do [method={%v},pattern={%v}]", requestMethod, pattern)
}

// Function callback paired with GET Method and URL-matching pattern.
func (api *API) Get(pattern string, fn interface{}) {
	api.Do("GET", pattern, fn)
}

// Function callback paired with PATH Method and URL-matching pattern.
func (api *API) Patch(pattern string, fn interface{}) {
	api.Do("PATCH", pattern, fn)
}

// Function callback paired with OPTIONS Method and URL-matching pattern.
func (api *API) Options(pattern string, fn interface{}) {
	api.Do("OPTIONS", pattern, fn)
}

// Function callback paired with HEAD Method and URL-matching pattern.
func (api *API) Head(pattern string, fn interface{}) {
	api.Do("HEAD", pattern, fn)
}

// Function callback paired with POST Method and URL-matching pattern.
func (api *API) Post(pattern string, fn interface{}) {
	api.Do("POST", pattern, fn)
}

// Function callback paired with LINK Method and URL-matching pattern.
func (api *API) Link(pattern string, fn interface{}) {
	api.Do("LINK", pattern, fn)
}

// Function callback paired with UNLINK Method and URL-matching pattern.
func (api *API) Unlink(pattern string, fn interface{}) {
	api.Do("UNLINK", pattern, fn)
}

// Function callback paired with PUT Method and URL-matching pattern.
func (api *API) Put(pattern string, fn interface{}) {
	api.Do("PUT", pattern, fn)
}

// Function callback paired with DELETE Method and URL-matching pattern.
func (api *API) Delete(fn interface{}, pattern string) {
	api.Do("DELETE", pattern, fn)
}

// Function callback paired with a set of URL-matching pattern.
func (api *API) addHandler(method string, handler http.HandlerFunc, pattern string) {
	api.logger.Debugf(" Add Handle Func [pattern={%v}]", pattern)
	pathChain := api.chain.Copy()
	pathChain.Target = handler
	handlerFunc := pathChain.dispatchRequestHandler()
	api.router.Add(pattern, method, handlerFunc)
}

//Implements HandlerFunc
func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, _ := api.mux.Handler(r)
	handler.ServeHTTP(w, r)
}

func (api *API) HandleFunc() {
	api.mux.HandleFunc("/", api.router.Handler(api.logger))
	api.router.OpsFriendlyLog(api.logger)
}

// Start causes the API to begin serving requests on the given port.
func (api *API) Start(port int) error {
	api.HandleFunc()
	portString := fmt.Sprintf(":%d", port)

	err := http.ListenAndServe(portString, api.mux)
	if err != nil {
		api.logger.Errorf(" API could not start at port %d \n", port)
		return err
	}
	return nil
}
