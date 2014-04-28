package pastis

import (
	"net/http"
)

//A simple admin REST resource 
type AdminResource struct {
}

//Simply GETs request and returns OK response
func (api AdminResource) Get() (int, interface{}) {
	return http.StatusOK, nil
}
