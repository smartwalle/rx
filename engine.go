package rx

import (
	"net/http"
	"sync"
)

var (
	default404Body = []byte("404 page not found")
	default405Body = []byte("405 method not allowed")
)

type Engine struct {
	*RouterGroup
	pool sync.Pool

	allNoRoute []HandlerFunc
	noRoute    []HandlerFunc
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

func (this *Engine) Use(handlers ...HandlerFunc) Router {
	this.RouterGroup.Use(handlers...)
	this.rebuild404Handlers()
	return this
}

func (this *Engine) NoRoute(handlers ...HandlerFunc) {
	this.noRoute = handlers
	this.rebuild404Handlers()
}

func (this *Engine) rebuild404Handlers() {
	this.allNoRoute = this.combineHandlers(this.noRoute)
}

func (this *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var c = this.pool.Get().(*Context)
	c.reset()
	c.Writer = w
	c.Request = req

	this.handleHTTPRequest(c)

	this.pool.Put(c)
}

func (this *Engine) handleHTTPRequest(c *Context) {
	var method = c.Request.Method
	var path = cleanPath(c.Request.URL.Path)

	var tree = this.RouterGroup.trees[method]
	if tree != nil {
		var nodes = tree.find(path, false)
		if len(nodes) > 0 {
			var node = nodes[0]
			if ok := this.exec(c, path, node); ok {
				return
			}
		} else {
			nodes = tree.find(path, true)
			for _, node := range nodes {
				if ok := this.exec(c, path, node); ok {
					return
				}
			}
		}
	}

	c.handlers = this.allNoRoute
	this.handleError(c, http.StatusNotFound, default404Body)
}

func (this *Engine) exec(c *Context, path string, node *pathNode) bool {
	if params, ok := node.match(path); ok {
		c.params = params
		c.handlers = node.handlers
		c.Next()
		return true
	}
	return false
}

func (this *Engine) handleError(c *Context, status int, content []byte) {
	c.Next()
	c.Writer.WriteHeader(status)
	c.Writer.Write(content)
}
