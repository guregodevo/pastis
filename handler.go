package pastis

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"log"
	"net/http"
	"net/url"
)


// An API manages a group of resources by routing to requests
// to the correct method on a matching resource and marshalling
// the returned data to JSON for the HTTP response.
//
// You can instantiate multiple APIs on separate ports. Each API
// will manage its own set of resources.
type API struct {
	mux    *http.ServeMux
	chain  *FilterChain
	router *Router
}

// NewAPI allocates and returns a new API.
func NewAPI() *API {
	return &API{chain: &FilterChain{[]Filter{}, 0, nil}, mux : http.NewServeMux(), router: NewRouter()}
}

func ErrorResponse(err error) interface{} {
	return map[string]string{"error": err.Error()}
}

//Return an instance of http.HandlerFunc built from  a pair of request method and a callback value.
//The first callback input parameter is the set of URL query and path parameters.
//The second callback input parameter is the unmarshalled JSON body recieved from the request (if it exists).
func handleMethodCall(urlValues url.Values, request *http.Request, methodRef reflect.Value) (int, interface{}) {
	log.Println("DEBUG: handleMethodCall ", request.Method)
	if methodRef.Kind() == reflect.Invalid {
		return http.StatusNotImplemented, nil
	}

	methodType := methodRef.Type()
	methodArgSize := methodRef.Type().NumIn()

	if methodArgSize == 0 {
		log.Println("ERROR: method %v cannot have 0 argument", methodRef)
		return http.StatusNotImplemented, nil
	}

	if methodArgSize >= 3 {
		log.Println("ERROR: method %v cannot have more than 2 arguments", methodRef)
		return http.StatusNotImplemented, nil
	}

	valueOfUrlValues := reflect.ValueOf(urlValues)
	methodParameterValues := []reflect.Value{valueOfUrlValues}

	if methodArgSize == 1 {
		return handleReturn(methodRef, methodParameterValues)
	}

	expectedJSONType := methodType.In(1)

	expectedJSONValue := reflect.New(expectedJSONType)

	jsonInterface := expectedJSONValue.Interface()

	dec := json.NewDecoder(request.Body)
	for {
		if err := dec.Decode(jsonInterface); err == io.EOF {
			break
		} else if err != nil {
			log.Println("ERROR: unable to decode json blob. Check whether parameter type matches json type. \n")
			return http.StatusNotImplemented, nil
		}
	}

	jsonValue := reflect.ValueOf(jsonInterface) //.Elem()
	jsonValueType := jsonValue.Elem().Type()

	if expectedJSONType.Kind() != jsonValue.Elem().Type().Kind() {
		log.Println("ERROR: Unexpected JSON format. Should be of type '", expectedJSONType.Kind(), "' instead of ", jsonValueType.Kind())
		return http.StatusNotImplemented, nil
	} else if expectedJSONType != jsonValue.Type() {
		methodParameterValues = append(methodParameterValues, jsonValue.Elem())
	} else {
		log.Println("ERROR: Parameter type mismatches json type. Expected JSON format. ", expectedJSONType)
		return http.StatusNotImplemented, nil
	}
	return handleReturn(methodRef, methodParameterValues)
}

//Return the array of Value converted into a tuple (int, interface {} )
func handleReturn(methodRef reflect.Value, methodParameterValues []reflect.Value) (int, interface{}) {
	responseValues := methodRef.Call(methodParameterValues)
	if len(responseValues) != 2 {
		log.Println("ERROR: method %v does not return expected response (int, interface{}).", methodRef)
		return http.StatusNotImplemented, nil
	}
	//TODO Fix int conversion
	return int(responseValues[0].Int()), responseValues[1].Interface()
}

//Return an instance of http.HandlerFunc built from  a pair of request method and a callback fn.
//The first callback input parameter is the set of URL query and path parameters.
//The second callback input parameter is the unmarshalled JSON body recieved from the request (if it exists).
func (api *API) methodHandler(pattern string, requestMethod string, fn reflect.Value) http.HandlerFunc {
	return func(rw http.ResponseWriter, request *http.Request) {
		//params := api.extractParams(pattern, request, request.Form)
				
		code, data := handleMethodCall(request.Form, request, fn)
		
		handlerFuncReturn(code, data, rw)
	}
}

//Utility method writing status code and data to the given response 
func handlerFuncReturn(code int, data interface{}, rw http.ResponseWriter) {
	log.Printf("DEBUG: handlerFuncReturn %v", code)

	content, err := json.Marshal(data)
	if err != nil {
		log.Printf("ERROR: handlerFuncReturn could not marshall content [%v]", data)
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
// requests that match one of the given paths to the matching HTTP
// method on the resource.
func (api *API) AddResource(resource interface{}, pattern string) {
	methods := []string {"GET","PUT","POST","PATCH","DELETE","OPTIONS"}
	for _, requestMethod := range methods {
		methodRef := reflect.ValueOf(resource).MethodByName(requestMethod)
		if methodRef.Kind() != reflect.Invalid {
			handler := api.methodHandler(pattern, requestMethod, methodRef)
			api.addHandler(requestMethod, handler, pattern)
			log.Printf("DEBUG: Added Resource [method={%v},pattern={%v}]", requestMethod, pattern)
		}
	}	
}

// Function callback paired with a request Method and URL-matching pattern. 
func (api *API) Do(requestMethod string, fn interface{}, pattern string) {
	handler := api.methodHandler(pattern, requestMethod, reflect.ValueOf(fn))
	api.addHandler(requestMethod, handler, pattern)
	log.Printf("DEBUG: Added Do [method={%v},pattern={%v}]", requestMethod, pattern)
}

// Function callback paired with GET Method and URL-matching pattern. 
func (api *API) Get(fn interface{}, pattern string) {
	api.Do("GET", fn, pattern)
}

// Function callback paired with PATH Method and URL-matching pattern. 
func (api *API) Patch(fn interface{}, pattern string) {
	api.Do("PATCH", fn, pattern)
}

// Function callback paired with OPTIONS Method and URL-matching pattern. 
func (api *API) Options(fn interface{}, pattern string) {
	api.Do("OPTIONS", fn, pattern)
}

// Function callback paired with HEAD Method and URL-matching pattern. 
func (api *API) Head(fn interface{}, pattern string) {
	api.Do("HEAD", fn, pattern)
}

// Function callback paired with POST Method and URL-matching pattern. 
func (api *API) Post(fn interface{}, pattern string) {
	api.Do("POST", fn, pattern)
}

// Function callback paired with PUT Method and URL-matching pattern. 
func (api *API) Put(fn interface{}, pattern string) {
	api.Do("PUT", fn, pattern)
}

// Function callback paired with DELETE Method and URL-matching pattern. 
func (api *API) Delete(fn interface{}, pattern string) {
	api.Do("DELETE", fn, pattern)
}

// Function callback paired with a set of URL-matching pattern. 
func (api *API) addHandler(method string, handler http.HandlerFunc, pattern string) {
	log.Printf("DEBUG: Add Handle Func [pattern={%v}]", pattern)
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
	//TODO Handle multiple patterns
	api.mux.HandleFunc("/", api.router.Handler())
}

// Start causes the API to begin serving requests on the given port.
func (api *API) Start(port int) error {
	//if api.mux == nil {
	//	return errors.New("You must add at least one resource to this API.")
	//}
	api.HandleFunc()
	portString := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(portString, api.mux)
}
