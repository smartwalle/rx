package rx

import (
	"fmt"
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
	pool  sync.Pool
	trees methodTrees

	allNoRoute HandlerChain
	noRoute    HandlerChain
}

func New() *Engine {
	var e = &Engine{}
	e.RouterGroup = newRouterGroup()
	e.RouterGroup.isRoot = true
	e.RouterGroup.engine = e
	e.pool.New = func() interface{} {
		return newContext()
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

func (this *Engine) addRoute(method, path string, handlers HandlerChain) {
	asset(method != "", "HTTP method can not be empty")
	asset(path[0] == '/', "path must begin with '/'")
	asset(len(handlers) > 0, "there must be at least one handler")

	var root = this.trees.get(method)
	if root == nil {
		root = &node{}
		root.fullPath = "/"
		this.trees = append(this.trees, methodTree{method: method, root: root})
	}
	root.addRoute(path, handlers)

	logger.Output(3, fmt.Sprintf("%-8s %-30s --> %s (%d handlers)\n", method, path, nameOfFunction(handlers.Last()), handlers.Len()))
}

func (this *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var c = this.pool.Get().(*Context)
	c.reset(w, req)

	this.handleHTTPRequest(c)

	this.pool.Put(c)
}

func (this *Engine) handleHTTPRequest(c *Context) {
	var method = c.Request.Method
	var path = CleanPath(c.Request.URL.Path)

	for i := 0; i < len(this.trees); i++ {
		if this.trees[i].method != method {
			continue
		}

		var root = this.trees[i].root
		value := root.getValue(path, c.params, true)
		if value.handlers != nil {
			c.handlers = value.handlers
			c.params = value.params
			c.Next()
			c.Writer.WriteHeaderNow()
			return
		}
	}

	// 匹配失败，返回 404 错误
	c.handlers = this.allNoRoute
	this.handleError(c, http.StatusNotFound, default404Body)
}

func (this *Engine) handleError(c *Context, status int, body []byte) {
	var w = c.Writer
	w.WriteHeader(status)

	c.Next()

	if w.Written() {
		return
	}

	w.Write(body)
}
