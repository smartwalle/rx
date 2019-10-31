package rx

import (
	"fmt"
	"net/http"
	"testing"
)

func TestRouterGroup_Find(t *testing.T) {
	var r = newRouterGroup()
	r.GET("/", func(c *Context) {})
	r.GET("/t1", func(c *Context) {})
	r.GET("/t1/h1", func(c *Context) {})
	r.GET("/t1/h2", func(c *Context) {})
	r.GET("/t2/h1", func(c *Context) {})
	r.GET("/t2/h2", func(c *Context) {})
	r.GET("/t4/", func(c *Context) {})

	var tests = []struct {
		method     string
		path       string
		numOfNodes int
	}{
		{method: http.MethodGet, path: "", numOfNodes: 1},
		{method: http.MethodGet, path: "/", numOfNodes: 1},
		{method: http.MethodGet, path: "//", numOfNodes: 1},
		{method: http.MethodGet, path: "/t1", numOfNodes: 1},
		{method: http.MethodGet, path: "/t1/", numOfNodes: 1},
		{method: http.MethodGet, path: "/t1/h1", numOfNodes: 1},
		{method: http.MethodGet, path: "/t1/h2", numOfNodes: 1},
		{method: http.MethodGet, path: "/t2/h1", numOfNodes: 1},
		{method: http.MethodGet, path: "/t2/h2", numOfNodes: 1},
		{method: http.MethodGet, path: "/t2/h1/", numOfNodes: 1},
		{method: http.MethodGet, path: "/t2/h2/", numOfNodes: 1},
		{method: http.MethodGet, path: "/t4", numOfNodes: 1},
		{method: http.MethodGet, path: "/t4/", numOfNodes: 1},

		{method: http.MethodGet, path: "/t11", numOfNodes: 0},
		{method: http.MethodGet, path: "/t2", numOfNodes: 0},
		{method: http.MethodGet, path: "/t2/", numOfNodes: 0},
		{method: http.MethodGet, path: "/t3", numOfNodes: 0},
		{method: http.MethodGet, path: "/t3/h1", numOfNodes: 0},
		{method: http.MethodGet, path: "/t3/h1/", numOfNodes: 0},
		{method: http.MethodGet, path: "/t5", numOfNodes: 0},
	}

	fmt.Println(tests)

	//for _, test := range tests {
	//	if e := r.find(test.method, CleanPath(test.path), false, nil); len(e) != test.numOfNodes {
	//		t.Errorf("%s - %s 的匹配结果应该为 %d, 实际为 %d", test.method, test.path, test.numOfNodes, len(e))
	//	}
	//}
}
