package pastis

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"encoding/json"
	"io/ioutil"
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

func assert_HTTP_Response(t *testing.T, res *http.Response, expectedStatusCode int, expectedResponsebody interface{}) {
	expect(t, res.StatusCode, expectedStatusCode)
	var body []byte
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	var f Foo
	err = json.Unmarshal(body, &f)
	if err != nil {
		log.Fatal(err)
	}
	expect(t, f, expectedResponsebody)
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
	return http.StatusOK, Foo{"name", 1}
}

func Test_Pastis_Resource_Handler(t *testing.T) {
	resource := new(FooResource)
	p := NewAPI()
	p.AddResource(resource, "/foo")

	ts := httptest.NewServer(p)
	defer ts.Close()

	url := ts.URL + "/foo"
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	assert_HTTP_Response(t, res, http.StatusOK, Foo{"name", 1})
}

func Test_Pastis_Action_Handler(t *testing.T) {
	p := NewAPI()
	p.Do("GET", func(vals url.Values) (int, interface{}) {
		return http.StatusOK, Foo{"name", 1}
	}, "/foo")

	ts := httptest.NewServer(p)
	defer ts.Close()

	url := ts.URL + "/foo"
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	assert_HTTP_Response(t, res, http.StatusOK, Foo{"name", 1})
}

func Test_Pastis_Action_Having_Input_Handler(t *testing.T) {
	p := NewAPI()
	p.Do("POST", func(vals url.Values, input Foo) (int, interface{}) {
		return http.StatusOK, input
	}, "/foo")

	ts := httptest.NewServer(p)
	defer ts.Close()

	foo := Foo{"postedName", 1}
	buf, _ := json.Marshal(foo)
	body := bytes.NewBuffer(buf)

	url := ts.URL + "/foo"
	res, err := http.Post(url, "application/json", body)
	if err != nil {
		log.Fatal(err)
	}
	assert_HTTP_Response(t, res, http.StatusOK, foo)
}

func Test_Pastis_POST_Having_Input_Handler(t *testing.T) {
	p := NewAPI()
	p.Post(func(vals url.Values, input Foo) (int, interface{}) {
		return http.StatusOK, input
	}, "/foo")

	ts := httptest.NewServer(p)
	defer ts.Close()

	foo := Foo{"postedName", 1}
	buf, _ := json.Marshal(foo)
	body := bytes.NewBuffer(buf)

	url := ts.URL + "/foo"
	res, err := http.Post(url, "application/json", body)
	if err != nil {
		log.Fatal(err)
	}
	assert_HTTP_Response(t, res, http.StatusOK, foo)
}