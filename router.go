package pastis

import (
 "net/http"
 "regexp"
 "fmt"
)

type Router struct {
	handlers map[string] map[string] http.HandlerFunc	
}

//Prints out the routes
func (router *Router) OpsFriendLog(logger *Logger) {
	fmt.Print("API Routes \n")
	log := make(map[string][]string)

	for method, _ := range router.handlers {
		for pattern, _ := range router.handlers[method] {
			log[pattern] = []string { method }
		}
	}
	for pattern := range log {
		for _, method := range log[pattern] {
			fmt.Printf(" %v %s \n", method, pattern)
		}
	}	
}

func NewRouter() *Router {
	hs := make(map[string]map[string]http.HandlerFunc)
	return &Router{ hs }
}

func (router *Router) Add(pattern string, method string, handler http.HandlerFunc) {
	if router.handlers[method] == nil {
		router.handlers[method] = make(map[string] http.HandlerFunc)
	}
	router.handlers[method][pattern] = handler
}

//Build a regex based on the initial pattern 
func Regexp(pattern string) *regexp.Regexp {
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

func (router *Router) Handler(logger *Logger) http.HandlerFunc {
	return func(rw http.ResponseWriter, request *http.Request) {
		logger.Debugf("routing [request=%v]...", request)

		if request.ParseForm() != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		handlersForPattern := router.handlers[request.Method]

		for pattern := range handlersForPattern {
			ok, params := Match(Regexp(pattern), request.URL.Path)
			if (ok) {
				logger.Debugf("Extracting params : URL [%s] | Pattern [%s] \n", request.URL.Path, pattern)
				for key, _ := range params {
					request.Form.Set(key,params[key])
				}	
				handlersForPattern[pattern](rw, request)
				return		
			}
		}
		logger.Debugf("No handler found for [method=%s,url=%v] ", request.Method, request.URL.Path)
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}