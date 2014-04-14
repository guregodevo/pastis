package pastis

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"log"
	"net/http"
	"net/url"
	"regexp"
)

// An API manages a group of resources by routing to requests
// to the correct method on a matching resource and marshalling
// the returned data to JSON for the HTTP response.
//
// You can instantiate multiple APIs on separate ports. Each API
// will manage its own set of resources.
type API struct {
	mux   *http.ServeMux
	chain *FilterChain
}

// NewAPI allocates and returns a new API.
func NewAPI() *API {
	return &API{chain: &FilterChain{[]Filter{}, 0, nil}}
}

func ErrorResponse(err error) interface{} {
	return map[string]string{"error": err.Error()}
}

//Invokes the resource method given the HTTP Method.
//When a request contains a JSON blob, the resource method recieves its content decoded.
//The JSON blob is unmarshalled and converted to the type of the second parameter.
func handleResourceCall(urlValues url.Values, request *http.Request, resource interface{}) (int, interface{}) {
	log.Println("DEBUG: handleResourceCall ", request.Method)
	methodRef := reflect.ValueOf(resource).MethodByName(request.Method)
	return handleMethodCall(urlValues, request, methodRef)
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
		log.Printf("DEBUG: methodHandler [pattern=%v,request=%v] ", pattern, request)

		if request.ParseForm() != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		if request.Method != requestMethod {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		params := api.extractParams(pattern, request, request.Form)
				
		code, data := handleMethodCall(params, request, fn)
		
		handlerFuncReturn(code, data, rw)
	}
}

//Return an instance of http.HandlerFunc built from a resource.
//A resource must implement GET, POST, DELETE, PUT or PATCH method having one or two parameters.
//The first parameter is the set of URL query and path parameters.
//The second parameter is the JSON blob recieved as a request body (if it exists).
func (api *API) resourceHandler(pattern string, resource interface{}) http.HandlerFunc {
	return func(rw http.ResponseWriter, request *http.Request) {
		log.Printf("DEBUG: resourceHandler [pattern=%v,request=%v] ", pattern, request)

		if request.ParseForm() != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		
		params := api.extractParams(pattern, request, request.Form)
		
		code, data := handleResourceCall(params, request, resource)
		handlerFuncReturn(code, data, rw)
	}
}

//Build a regex based on the initial pattern 
func (api *API) regexp(pattern string) *regexp.Regexp {
	r := regexp.MustCompile(`:[^/#?()\.\\]+`)
	pattern = r.ReplaceAllStringFunc(pattern, func(m string) string {
		return fmt.Sprintf(`(?P<%s>[^/#?]+)`, m[1:])
	})
	r2 := regexp.MustCompile(`\*\*`)
	var index int
	pattern = r2.ReplaceAllStringFunc(pattern, func(m string) string {
		index++
		return fmt.Sprintf(`(?P<_%d>[^#?]*)`, index)
	})
	pattern += `\/?`
	return regexp.MustCompile(pattern)
}

func (api *API) extractParams(pattern string, request *http.Request, urlValues url.Values) url.Values {
	log.Printf("Extract params : URL [%s] | Pattern [%s] \n", request.URL.Path, pattern)
	ok, params := api.Match(api.regexp(pattern), request.URL.Path)
	if (ok) {
		fmt.Println("Expression regular matches")
		for key, _ := range params {
			fmt.Println(key)
			urlValues.Set(key,params[key])
		}	
	}
	return urlValues
}

// HandlerPath returns the Server path 
func HandlerPath(pattern string) string {
	reg := regexp.MustCompile("^/*[^:]*")
	matches := reg.FindString(pattern)
	if len(matches) > 0 {
		return matches
	}
	return pattern
}

// URLWith returns the url pattern replacing the parameters for its values
func ReplaceParametersWith(pattern string, str string) string {
	re := regexp.MustCompile("(?P<first>[a-zA-Z]+) (?P<last>[a-zA-Z]+)")
	fmt.Println(re.MatchString("Alan Turing"))
	fmt.Printf("%q\n", re.SubexpNames())

	reg := regexp.MustCompile(`:[^/#?()\.\\]+`)
	url := reg.ReplaceAllStringFunc(pattern, func(m string) string {
		log.Printf("Replacing [%s]", m)
		val := str
		return fmt.Sprintf(`%v`, val)
	})
	log.Printf("Replaced parameters of Pattern [%s] : now [%s]", pattern, url)
	return url	
}

func (api *API) Match(r *regexp.Regexp, path string) (bool, map[string]string) {
	matches := r.FindStringSubmatch(path)
	if len(matches) > 0 && matches[0] == path {
		params := make(map[string]string)
		for i, name := range r.SubexpNames() {
			if len(name) > 0 {
				params[name] = matches[i]
			}
		}
		return true, params
	}
	return false, nil
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
	if api.mux == nil {
		api.mux = http.NewServeMux()
	}
	handler := api.resourceHandler(pattern, resource)
	api.addHandler(handler, HandlerPath(pattern))
}

// Function callback paired with a request Method and URL-matching pattern. 
func (api *API) Do(requestMethod string, fn interface{}, pattern string) {
	if api.mux == nil {
		api.mux = http.NewServeMux()
	}
	handler := api.methodHandler(pattern, requestMethod, reflect.ValueOf(fn))
	api.addHandler(handler, HandlerPath(pattern))
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
func (api *API) addHandler(handler http.HandlerFunc, pattern string) {
	log.Printf("DEBUG: Handle Func [pattern={%v}]", pattern)
	pathChain := api.chain.Copy()
	pathChain.Target = handler
	api.mux.HandleFunc(pattern, pathChain.dispatchRequestHandler())
}

//Implements HandlerFunc
func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if api.mux == nil {
		log.Panic(errors.New("You must add at least one resource to this API."))
	}
	handler, _ := api.mux.Handler(r)
	handler.ServeHTTP(w, r)
}

// Start causes the API to begin serving requests on the given port.
func (api *API) Start(port int) error {
	if api.mux == nil {
		return errors.New("You must add at least one resource to this API.")
	}
	portString := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(portString, api.mux)
}
