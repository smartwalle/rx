package rx

import (
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
		path       string
		numOfNodes int
	}{
		{path: "/", numOfNodes: 1},
		{path: "/t1", numOfNodes: 1},
		{path: "/t1/h1", numOfNodes: 1},
		{path: "/t1/h2", numOfNodes: 1},
		{path: "/t2/h1", numOfNodes: 1},
		{path: "/t2/h2", numOfNodes: 1},
		{path: "/t4", numOfNodes: 1},

		{path: "", numOfNodes: 0},
		{path: "//", numOfNodes: 0},
		{path: "/t1/", numOfNodes: 0},
		{path: "/t2", numOfNodes: 0},
		{path: "/t11", numOfNodes: 0},
		{path: "/t3", numOfNodes: 0},
		{path: "/t3/h1", numOfNodes: 0},
		{path: "/t3/h1/", numOfNodes: 0},
		{path: "/t5", numOfNodes: 0},
	}

	for _, test := range tests {
		treeFindTest(t, tree, test.path, false, test.numOfNodes)
	}
}

func TestTree_FindRegex(t *testing.T) {
	var handlers = HandlerChain{}
	handlers = append(handlers, func(c *Context) {})

	var tree = newMethodTree("GET")
	tree.add("/", handlers)
	tree.add("/user/:id", handlers)
	tree.add("/user/:id/point", handlers)
	tree.add("/user/list", handlers)
	tree.add("/order/:id", handlers)
	tree.add("/order/:id/*action", handlers)
	tree.add("/point/:id", handlers)
	tree.add("/point/list", handlers)

	var tests = []struct {
		path       string
		numOfNodes int
	}{
		{path: "/user", numOfNodes: 2},      // 匹配到 /user/:id、/user/:id/point
		{path: "/user/1", numOfNodes: 2},    // 匹配到 /user/:id、/user/:id/point
		{path: "/order/1", numOfNodes: 2},   // 匹配到 /order/:id、/order/:id/*action
		{path: "/point/1", numOfNodes: 1},   // 匹配到 /point/:id
		{path: "/user/list", numOfNodes: 0}, // 非正则
	}

	for _, test := range tests {
		treeFindTest(t, tree, test.path, true, test.numOfNodes)
	}
}

func TestTree_Clean(t *testing.T) {
	var handlers = HandlerChain{}
	handlers = append(handlers, func(c *Context) {})

	var paths []string
	paths = append(paths, "/")
	paths = append(paths, "/t1")
	paths = append(paths, "/t1/h1")
	paths = append(paths, "/t1/h2")
	paths = append(paths, "/t2/h1")
	paths = append(paths, "/t2/h2")
	paths = append(paths, "/t4")

	var tree = newMethodTree("GET")
	for _, path := range paths {
		tree.add(path, handlers)
	}

	for _, path := range paths {
		tree.clean(path)
		treeFindTest(t, tree, path, false, 0)
	}
}

func treeFindTest(t *testing.T, tree *methodTree, path string, isRegex bool, numOfNodes int) {
	if nodes := tree.find(path, isRegex, nil); len(nodes) != numOfNodes {
		t.Errorf("%s 的匹配结果应该为 %d, 实际为 %d\n", path, numOfNodes, len(nodes))
	}
}
