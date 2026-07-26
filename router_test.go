package rx_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/smartwalle/rx"
)

type testRouter struct {
	routes []testRoute
}

type testRoute struct {
	regexp *regexp.Regexp
	route  rx.Route
}

func newTestRouter() *testRouter {
	return &testRouter{}
}

func (r *testRouter) Add(pattern string, targets []string, opts ...rx.Option) error {
	route, err := rx.NewProxyRoute(pattern, targets, opts...)
	if err != nil {
		return err
	}
	return r.AddRoute(route)
}

func (r *testRouter) AddRoute(route rx.Route) error {
	nRegexp, err := regexp.Compile(route.Pattern())
	if err != nil {
		return err
	}
	r.routes = append(r.routes, testRoute{regexp: nRegexp, route: route})
	return nil
}

func (r *testRouter) Match(req *http.Request) (rx.Route, error) {
	for _, route := range r.routes {
		if route.regexp.MatchString(req.URL.Path) {
			return route.route, nil
		}
	}
	return nil, nil
}

func TestRouter_Match(t *testing.T) {
	var provider = newTestRouter()
	// 要想能够正常匹配 /user, 需要在添加 /u 之前添加 /user
	provider.Add("/user", []string{"http://127.0.0.1:9920", "http://127.0.0.1:9921"})
	provider.Add("/u", []string{"http://127.0.0.1:9910", "http://127.0.0.1:9911"})

	var tests = []struct {
		path    string
		pattern string
	}{
		{
			path:    "/u",
			pattern: "/u",
		},
		{
			path:    "/u/u",
			pattern: "/u",
		},
		{
			path:    "/us",
			pattern: "/u",
		},
		{
			path:    "/us/u",
			pattern: "/u",
		},
		{
			path:    "/use",
			pattern: "/u",
		},
		{
			path:    "/use/u",
			pattern: "/u",
		},
		{
			path:    "/user",
			pattern: "/user",
		},
		{
			path:    "/user/u",
			pattern: "/user",
		},
	}

	for _, test := range tests {
		var req = httptest.NewRequest("GET", "http://127.0.0.1"+test.path, nil)
		var route, _ = provider.Match(req)
		if route.Pattern() != test.pattern {
			t.Fatalf("期望: %s，实际：%s \n", test.pattern, route.Pattern())
		}
	}

}
