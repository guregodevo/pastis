package pastis

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_Pastis_AdminResource_Handler(t *testing.T) {
	resource := new(AdminResource)
	p := NewAPI()
	p.AddResource("/ping", resource)
	p.HandleFunc()

	ts := httptest.NewServer(p)
	defer ts.Close()

	url := ts.URL + "/ping"
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	expect(t, res.StatusCode, http.StatusOK)
}
