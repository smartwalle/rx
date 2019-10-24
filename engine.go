package rx

import (
	"fmt"
	"net/http"
)

type Engine struct {
	*RouterGroup
}

func New() *Engine {
	var e = &Engine{}
	e.RouterGroup = newRouterGroup()
	e.RouterGroup.isRoot = true
	e.RouterGroup.engine = e
	return e
}

func (this *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var path = cleanPath(req.URL.Path)

	var tree = this.RouterGroup.trees[req.Method]
	if tree == nil {
		// TODO method not allowed
		return
	}

	var nodes = tree.Find(path, false)
	if len(nodes) > 0 {
		var node = nodes[0]
		if node.path == path && len(node.handlers) > 0 {
			this.handle(node, w, req)
			return
		}
	} else {
		// TODO regex
	}

	fmt.Println("bad request")

	// TODO not found
}

func (this *Engine) handle(node *Node, w http.ResponseWriter, req *http.Request) {
	var c = &Context{}
	c.reset()
	c.Request = req
	c.Writer = w
	c.handlers = node.handlers
	c.Next()
}
