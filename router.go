package pastis

import "net/http"
import "log"
//import "net/url"
import "fmt"
import "regexp"

type Router struct {
	handlers map[string] map[string] http.HandlerFunc	
}

func NewRouter() *Router {
	hs := make(map[string]map[string]http.HandlerFunc)
	return &Router{ hs }
}

func (router *Router) Add(pattern string, method string, handler http.HandlerFunc) error {
	if router.handlers[method] == nil {
		router.handlers[method] = make(map[string] http.HandlerFunc)
	}
	router.handlers[method][pattern] = handler
	return nil
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

func (router *Router) Handler() http.HandlerFunc {
	return func(rw http.ResponseWriter, request *http.Request) {
		log.Printf("DEBUG: router Handler [request=%v] ", request)

		if request.ParseForm() != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		handlersForPattern := router.handlers[request.Method]

		for pattern := range handlersForPattern {
			ok, params := Match(Regexp(pattern), request.URL.Path)
			if (ok) {
				log.Printf("Extract params : URL [%s] | Pattern [%s] \n", request.URL.Path, pattern)
				for key, _ := range params {
					request.Form.Set(key,params[key])
				}	
				handlersForPattern[pattern](rw, request)
				return		
			}
		}
		log.Printf("DEBUG: No handler for [method=%v,url=%v] ", request.Method, request.URL.Path)
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}


