package pastis

import (
	"log"
	"net/http"
	"strings"
)

const (
	HEADER_Allow_Methods                 = "GET,POST,DELETE,PUT,PATCH,HEAD"
	HEADER_Access_Control_Allow_Headers  = "Origin,Accept,Produce,Content-Type,X-Requested-With,Authorization,Token"
	HEADER_Origin                        = "Origin"
	HEADER_Allow_Credentials             = "Access-Control-Allow-Credentials"
	HEADER_Allow_Origin                  = "Access-Control-Allow-Origin"
	HEADER_AccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HEADER_Access_Control_Request_Method = "Access-Control-Request-Method"
)

//A Cross Origin Filter
func CORSFilter(rw http.ResponseWriter, request *http.Request, chain *FilterChain) {
	log.Printf("[HTTP] CORS Filtering %s %s\n", request.Method, request.URL)

	if origin := request.Header.Get(HEADER_Origin); len(origin) == 0 {
		chain.NextFilter(rw, request)
		return
	}

	if request.Method != "OPTIONS" {
		handleRequest(rw, request, chain)
		return
	}

	if controlRequestMethod := request.Header.Get(HEADER_Access_Control_Request_Method); controlRequestMethod == "" {
		handleRequest(rw, request, chain)
	} else {
		handlePreflight(rw, request, chain)
	}
}

func setAllowOriginHeader(rw http.ResponseWriter, request *http.Request) {
	origin := request.Header.Get(HEADER_Origin)
	rw.Header().Add(HEADER_Allow_Origin, origin)
}

func setAllowCredentialsHeader(rw http.ResponseWriter) {
	rw.Header().Add(HEADER_Allow_Credentials, "true")
}

func validateAccessControlRequestMethod(method string, allowedMethods []string) bool {
	for _, e := range allowedMethods {
		if e == method {
			return true
		}
	}
	return false
}

func validateAccessControlRequestHeader(header string) bool {
	for _, e := range strings.Split(HEADER_Access_Control_Allow_Headers, ",") {
		if strings.ToLower(e) == strings.ToLower(header) {
			return true
		}
	}
	return false
}

func handlePreflight(rw http.ResponseWriter, request *http.Request, chain *FilterChain) {
	if !validateAccessControlRequestMethod(request.Header.Get(HEADER_Access_Control_Request_Method), strings.Split(HEADER_Allow_Methods, ",")) {
		chain.NextFilter(rw, request)
		return
	}
	if controlReqHeaders := request.Header.Get(HEADER_AccessControlRequestHeaders); len(controlReqHeaders) > 0 {
		for _, e := range strings.Split(controlReqHeaders, ",") {
			if !validateAccessControlRequestHeader(strings.Trim(e, " ")) {
				chain.NextFilter(rw, request)
				return
			}
		}
	}
	setAllowOriginHeader(rw, request)
	rw.Header().Set("Access-Control-Allow-Methods", HEADER_Allow_Methods)
	rw.Header().Add("Access-Control-Allow-Headers", HEADER_Access_Control_Allow_Headers)
	setAllowCredentialsHeader(rw)
}

func handleRequest(rw http.ResponseWriter, request *http.Request, chain *FilterChain) {
	setAllowCredentialsHeader(rw)
	setAllowOriginHeader(rw, request)
	chain.NextFilter(rw, request)
}
