package pastis

import (
	"net/http"
)

type AdminResource struct {
}

func (api AdminResource) Get() (int, interface{}) {
	return http.StatusOK, nil
}

