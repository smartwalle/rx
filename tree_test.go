package rx

import (
	"fmt"
	"net/http"
	"testing"
)

func TestTree_Find(t *testing.T) {
	var handlers = HandlerChain{}
	handlers = append(handlers, func(c *Context) {})

	var tree = newMethodTree("GET")
	tree.add("/", handlers)
	tree.add("/t1", handlers)
	tree.add("/t1/h1", handlers)
	tree.add("/t1/h2", handlers)
	tree.add("/t2/h1", handlers)
	tree.add("/t2/h2", handlers)
	tree.add("/t4", handlers)
	tree.add("/t5", nil)

	var tests = []struct {
		method     string
		path       string
		numOfNodes int
	}{
		{method: http.MethodGet, path: "/", numOfNodes: 1},
		{method: http.MethodGet, path: "/t1", numOfNodes: 1},
		{method: http.MethodGet, path: "/t1/h1", numOfNodes: 1},
		{method: http.MethodGet, path: "/t1/h2", numOfNodes: 1},
		{method: http.MethodGet, path: "/t2/h1", numOfNodes: 1},
		{method: http.MethodGet, path: "/t2/h2", numOfNodes: 1},
		{method: http.MethodGet, path: "/t4", numOfNodes: 1},

		{method: http.MethodGet, path: "", numOfNodes: 0},
		{method: http.MethodGet, path: "//", numOfNodes: 0},
		{method: http.MethodGet, path: "/t1/", numOfNodes: 0},
		{method: http.MethodGet, path: "/t2", numOfNodes: 0},
		{method: http.MethodGet, path: "/t11", numOfNodes: 0},
		{method: http.MethodGet, path: "/t3", numOfNodes: 0},
		{method: http.MethodGet, path: "/t3/h1", numOfNodes: 0},
		{method: http.MethodGet, path: "/t3/h1/", numOfNodes: 0},
		{method: http.MethodGet, path: "/t5", numOfNodes: 0},
	}

	for _, test := range tests {
		if e := tree.find(test.path, false); len(e) != test.numOfNodes {
			t.Errorf("%s - %s 的匹配结果应该为 %d, 实际为 %d", test.method, test.path, test.numOfNodes, len(e))
		}
	}
}

func TestTree_FindRegex(t *testing.T) {

}

func TestTree_Clean(t *testing.T) {
	var handlers = HandlerChain{}
	handlers = append(handlers, func(c *Context) {})

	var tree = newMethodTree("GET")
	tree.add("/", handlers)
	tree.add("/t1", handlers)
	tree.add("/t1/h1", handlers)
	tree.add("/t1/h2", handlers)
	tree.add("/t2/h1", handlers)
	tree.add("/t2/h2", handlers)
	tree.add("/t4", handlers)

	tree.clean("/t2/h1")
	tree.clean("/t2/h2")
	tree.clean("/t1")
	tree.clean("/t1/h1")
	tree.clean("/t1/h2")
	tree.clean("/t4")
	tree.clean("/")

	fmt.Println("-------------------")
	tree.print()
}
