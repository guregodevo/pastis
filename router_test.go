package pastis

import (
	"testing"
)

func Test_Pastis_URLMatch(t *testing.T) {
	ok, params := Match(Regexp("/hello/:name"), "/hello/guregodevo")
	expect(t, ok, true)
	expect(t, params["name"], "guregodevo")
	
	ok, params = Match(Regexp("/hello/:name/:id"), "/hello/guregodevo/1234")
	expect(t, ok, true)
	expect(t, params["name"], "guregodevo")
	expect(t, params["id"], "1234")
	
	
}

func Test_Pastis_RegexpMatch(t *testing.T) {
	ok, params := Match(Regexp("^/comment/(?P<id>\\d+)$"), "/comment/123")
	expect(t, ok, true)
	expect(t, params["id"], "123")
}

