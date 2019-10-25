package rx

import (
	"fmt"
	"net/http"
	"sync"
)

type Engine struct {
	*RouterGroup
	pool sync.Pool
}

func New() *Engine {
	var e = &Engine{}
	e.RouterGroup = newRouterGroup()
	e.RouterGroup.isRoot = true
	e.RouterGroup.engine = e
	e.pool.New = func() interface{} {
		return &Context{}
	}
	return e
}

func (this *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var path = cleanPath(req.URL.Path)

	var tree = this.RouterGroup.trees[req.Method]
	if tree == nil {
		// TODO method not allowed
		return
	}

	var nodes = tree.find(path, false)
	if len(nodes) > 0 {
		var node = nodes[0]
		if ok := this.handleHttpRequest(node, path, w, req); ok {
			return
		}
	} else {
		nodes = tree.find(path, true)
		for _, node := range nodes {
			if ok := this.handleHttpRequest(node, path, w, req); ok {
				return
			}
		}
	}

	fmt.Println("bad request")

	// TODO not found
}

func (this *Engine) handleHttpRequest(node *pathNode, path string, w http.ResponseWriter, req *http.Request) bool {
	if len(node.handlers) > 0 {
		if params, ok := node.match(path); ok {
			var c = this.pool.Get().(*Context)
			c.reset()
			c.Request = req
			c.Writer = w
			c.handlers = node.handlers
			c.params = params
			c.Next()
			this.pool.Put(c)
			return true
		}
	}
	return false
}
