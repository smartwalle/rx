package rx

import (
	"net/http"
	"testing"
)

func TestNewRouter(t *testing.T) {
	var r = newRouterGroup()
	r.GET("/", func(writer http.ResponseWriter, request *http.Request) {})
	r.GET("/t1", func(writer http.ResponseWriter, request *http.Request) {})
	r.GET("/t1/h1", func(writer http.ResponseWriter, request *http.Request) {})
	r.GET("/t1/h2", func(writer http.ResponseWriter, request *http.Request) {})
	r.GET("/t2/h1", func(writer http.ResponseWriter, request *http.Request) {})
	r.GET("/t2/h2", func(writer http.ResponseWriter, request *http.Request) {})
	r.GET("/t4/", func(writer http.ResponseWriter, request *http.Request) {})

	var tests = []struct {
		m string
		p string
		c int
	}{
		{m: http.MethodGet, p: "/", c: 1},
		{m: http.MethodGet, p: "//", c: 1},
		{m: http.MethodGet, p: "/t1", c: 1},
		{m: http.MethodGet, p: "/t1/", c: 1},
		{m: http.MethodGet, p: "/t1/h1", c: 1},
		{m: http.MethodGet, p: "/t1/h2", c: 1},
		{m: http.MethodGet, p: "/t2/h1", c: 1},
		{m: http.MethodGet, p: "/t2/h2", c: 1},
		{m: http.MethodGet, p: "/t4", c: 1},
		{m: http.MethodGet, p: "/t4/", c: 1},
		{m: http.MethodGet, p: "/t2/h1/", c: 1},
		{m: http.MethodGet, p: "/t2/h2/", c: 1},

		{m: http.MethodGet, p: "/t11", c: 0},
		{m: http.MethodGet, p: "/t2", c: 0},
		{m: http.MethodGet, p: "/t2/", c: 0},
		{m: http.MethodGet, p: "/t3", c: 0},
		{m: http.MethodGet, p: "/t3/h1", c: 0},
		{m: http.MethodGet, p: "/t3/h1/", c: 0},
	}

	for _, test := range tests {
		if e := r.find(test.m, test.p, false); len(e) != test.c {
			t.Errorf("%s - %s 的匹配结果应该为 %d, 实际为 %d", test.m, test.p, test.c, len(e))
		}
	}

	r.Print()
}
