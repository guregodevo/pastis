package pastis

import (
	"fmt"
	"net/http"
	"regexp"
)

//Router is a struct consisting of a set of method paired with URL-matchin pattern where each pair is mapped to an handler function. 
type Router struct {
	handlers map[string]map[string]http.HandlerFunc
}

//Prints out the routes in a friendly manner
func (router *Router) OpsFriendlyLog(logger *Logger) {
	fmt.Print("API Routes \n")
	log := make(map[string][]string)

	for method, _ := range router.handlers {
		for pattern, _ := range router.handlers[method] {
			log[pattern] = []string{method}
		}
	}
	for pattern := range log {
		for _, method := range log[pattern] {
			fmt.Printf(" %v %s \n", method, pattern)
		}
	}
}

// NewRouter allocates and returns a new Router.
func NewRouter() *Router {
	hs := make(map[string]map[string]http.HandlerFunc)
	return &Router{hs}
}

// Add adds a route wichi consist of an URL-pattern matching, a method and an handler of type http.HandlerFunc
func (router *Router) Add(pattern string, method string, handler http.HandlerFunc) {
	if router.handlers[method] == nil {
		router.handlers[method] = make(map[string]http.HandlerFunc)
	}
	router.handlers[method][pattern] = handler
}

//Regex simply builds a more reliable regex based on the initial pattern
func Regexp(pattern string) *regexp.Regexp {
	r := regexp.MustCompile(`:[^/#?()\.\\]+`)
	pattern = r.ReplaceAllStringFunc(pattern, func(m string) string {
		return fmt.Sprintf(`(?P<%s>[^/#?]+)`, m[1:])
	})
	rr := regexp.MustCompile(`\*\*`)
	var index int
	pattern = rr.ReplaceAllStringFunc(pattern, func(m string) string {
		index++
		return fmt.Sprintf(`(?P<_%d>[^#?]*)`, index)
	})
	pattern += `\/?`
	return regexp.MustCompile(pattern)
}

//Match checks whether the given pat matches the given regular expresion. 
func Match(r *regexp.Regexp, path string) (bool, map[string]string) {
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

// HandlerPath returns the Server path
func HandlerPath(pattern string) string {
	reg := regexp.MustCompile("^/*[^:]*")
	matches := reg.FindString(pattern)
	if len(matches) > 0 {
		return matches
	}
	return pattern
}

//Handler returns an handler function of the API. 
//This handler is built from the set of routes that have been
//defined previously. 
func (router *Router) Handler(logger *Logger) http.HandlerFunc {
	return func(rw http.ResponseWriter, request *http.Request) {
		logger.Debugf("routing [request=%v]...", request)

		if request.ParseForm() != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		method := request.Method

		if method == "OPTIONS" &&  request.Header.Get(HEADER_Access_Control_Request_Method) != "" {
			method = request.Header.Get(HEADER_Access_Control_Request_Method)
			logger.Debugf("CORS negotiation initiaded: Routing to the Access control method [%v] ", method) 
		}	

		handlersForPattern := router.handlers[method]

		for pattern := range handlersForPattern {
			ok, params := Match(Regexp(pattern), request.URL.Path)
			if ok {
				logger.Debugf("Extracting params : URL [%s] | Pattern [%s] \n", request.URL.Path, pattern)
				for key, _ := range params {
					request.Form.Set(key, params[key])
				}
				handlersForPattern[pattern](rw, request)
				return
			}
		}	
		logger.Debugf("No handler found for [method=%s,url=%v] ", method, request.URL.Path)
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}
