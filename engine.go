package rx

import (
	"net/http"
	"sync"
)

var (
	default404Body = []byte("404 page not found")
	default405Body = []byte("405 method not allowed")
)

type HandlerFunc func(c *Context)

type HandlerChain []HandlerFunc

func (this HandlerChain) Last() HandlerFunc {
	if l := len(this); l > 0 {
		return this[l-1]
	}
	return nil
}

func (this HandlerChain) Len() int {
	return len(this)
}

type Engine struct {
	*RouterGroup
	pool sync.Pool

	allNoRoute HandlerChain
	noRoute    HandlerChain
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
	c.reset(w, req)

	this.handleHTTPRequest(c)

	this.pool.Put(c)
}

func (this *Engine) handleHTTPRequest(c *Context) {
	var method = c.Request.Method
	var path = cleanPath(c.Request.URL.Path)

	var tree = this.getTree(method)
	if tree != nil {
		// 先使用完整路径进行匹配
		var nodes = tree.find(path, false)
		if len(nodes) > 0 {
			var node = nodes[0]
			if ok := this.exec(c, path, node); ok {
				return
			}
		} else {
			// 完整路径匹配失败，则尝试正则匹配
			nodes = tree.find(path, true)
			for _, node := range nodes {
				if ok := this.exec(c, path, node); ok {
					return
				}
			}
		}
	}

	// 匹配失败，返回 404 错误
	c.handlers = this.allNoRoute
	this.handleError(c, http.StatusNotFound, default404Body)
}

func (this *Engine) exec(c *Context, path string, node *treeNode) bool {
	if params, ok := node.match(path); ok {
		c.params = params
		c.handlers = node.handlers
		c.Next()
		return true
	}
	return false
}

func (this *Engine) handleError(c *Context, status int, content []byte) {
	var w = c.Writer.(*responseWriter)
	w.WriteHeader(status)

	c.Next()

	if w.Written() {
		return
	}

	w.Write(content)
}
