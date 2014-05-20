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

func Test_Pastis_ComplexRegexpMatch(t *testing.T) {
	regexp := "^/dashboards/:dashboardid/chart/(?P<chartid>[0-9]*)$"
	ok, params := Match(Regexp(regexp), "/dashboards/1/chart/2")
	expect(t, ok, true)
	expect(t, params["chartid"], "2")
	expect(t, params["dashboardid"], "1")

	ok, params = Match(Regexp(regexp), "/dashboards/1/chart/")
	expect(t, ok, true)
	expect(t, params["chartid"], "")
	expect(t, params["dashboardid"], "1")
}

