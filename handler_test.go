package pastis

import (
	"net/http"
	"net/http/httptest"
	"log"
	//"fmt"
	"io/ioutil"
	"encoding/json"
	"net/url"
	"reflect"
	"testing"
)

/* Test Helpers */
func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func refute(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		t.Errorf("Did not expect %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func Test_NewAPI(t *testing.T) {
	m := NewAPI()
	if m == nil {
		t.Error("pastis.New() cannot return nil")
	}
}

func Test_Pastis_Run(t *testing.T) {
	// just test that Start doesn't bomb
	go NewAPI().Start(3000)
}

type FooResource struct {
}

type Foo struct {
		Name  string
		Order int
}

func (api FooResource) GET(vals url.Values) (int, interface{}) {
	return http.StatusOK, Foo {"name", 1}
}

func Test_Pastis_Handler(t *testing.T) {
	resource := new(FooResource)
	p := NewAPI()
	p.AddResource(resource, "/foo")
	
	ts := httptest.NewServer(p)
	defer ts.Close()

	url:= ts.URL + "/foo"
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	var body []byte
	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	} 
	var f Foo
	err = json.Unmarshal(body, &f)
	if err != nil {
		log.Fatal(err)
	}
	expect(t, res.StatusCode, http.StatusOK)
	expect(t, f,  Foo {"name", 1})
}