package rx

import (
	"fmt"
	"github.com/smartwalle/rx/balancer/roundrobin"
	"net/http"
	"net/url"
	"sync"
)

var (
	default404Body = []byte("404 page not found")
	default405Body = []byte("405 method not allowed")
)

type HandlerFunc func(c *Context)

type HandlersChain []HandlerFunc

func (this HandlersChain) Last() HandlerFunc {
	if l := len(this); l > 0 {
		return this[l-1]
	}
	return nil
}

func (this HandlersChain) Len() int {
	return len(this)
}

type Engine struct {
	*RouterGroup
	pool  sync.Pool
	trees methodTrees

	allNoRoute HandlersChain
	noRoute    HandlersChain

	allNoMethod HandlersChain
	noMethod    HandlersChain
}

func New() *Engine {
	var e = &Engine{}
	e.RouterGroup = newRouterGroup()
	e.RouterGroup.engine = e
	e.pool.New = func() interface{} {
		return newContext()
	}
	return e
}

func (this *Engine) Use(handlers ...HandlerFunc) Router {
	this.RouterGroup.Use(handlers...)
	this.rebuild404Handlers()
	this.rebuild405Handlers()
	return this
}

func (this *Engine) NoRoute(handlers ...HandlerFunc) {
	this.noRoute = handlers
	this.rebuild404Handlers()
}

func (this *Engine) NoMethod(handlers ...HandlerFunc) {
	this.noMethod = handlers
	this.rebuild405Handlers()
}

func (this *Engine) rebuild404Handlers() {
	this.allNoRoute = this.combineHandlers(this.noRoute)
}

func (this *Engine) rebuild405Handlers() {
	this.allNoMethod = this.combineHandlers(this.noMethod)
}

func (this *Engine) addRoute(method, path string, targets []string, handlers HandlersChain) {
	asset(method != "", "HTTP method can not be empty")
	asset(path[0] == '/', "path must begin with '/'")
	asset(len(targets) > 0, "there must be at least one target")

	var root = this.trees.get(method)
	if root == nil {
		root = &node{}
		root.fullPath = "/"
		this.trees = append(this.trees, methodTree{method: method, root: root})
	}

	var nTargets = make([]*url.URL, 0, len(targets))

	for _, target := range targets {
		var nURL, err = url.Parse(target)
		if err != nil {
			panic(err.Error())
		}
		nTargets = append(nTargets, nURL)
	}

	var balancer, err = (&roundrobin.Builder{}).Build(nTargets)
	if err != nil {
		panic(err.Error())
	}

	var route = &Route{}
	route.handlers = handlers
	route.balancer = balancer

	root.addRoute(path, route)

	logger.Output(3, fmt.Sprintf("%-8s %-30s --> %s (%d handlers)\n", method, path, nameOfFunction(handlers.Last()), handlers.Len()))
}

func (this *Engine) breakRoute(method, path string) {
	asset(method != "", "HTTP method can not be empty")
	asset(path[0] == '/', "path must begin with '/'")

	for i := 0; i < len(this.trees); i++ {
		if this.trees[i].method != method {
			continue
		}
		var root = this.trees[i].root
		var node = root.getNode(path)

		if node != nil {
			node.route = nil
		}
	}
}

func (this *Engine) existRoute(method, path string) bool {
	asset(method != "", "HTTP method can not be empty")
	asset(path[0] == '/', "path must begin with '/'")

	for i := 0; i < len(this.trees); i++ {
		if this.trees[i].method != method {
			continue
		}
		var root = this.trees[i].root
		var node = root.getNode(path)

		if node != nil && node.route != nil {
			return true
		}
	}
	return false
}

func (this *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var c = this.pool.Get().(*Context)
	c.reset(w, req)

	this.handleHTTPRequest(c)

	this.pool.Put(c)
}

func (this *Engine) handleHTTPRequest(c *Context) {
	var method = c.Request.Method
	var path = c.Request.URL.Path

	var ts = this.trees
	var tl = len(ts)
	for i := 0; i < tl; i++ {
		if ts[i].method != method {
			continue
		}

		var root = ts[i].root
		var value = root.getValue(path, c.params, false)
		if value.route != nil {
			c.handlers = value.route.handlers
			c.balancer = value.route.balancer
			c.params = value.params
			c.exec()
			return
		}
	}

	// 匹配 405 错误
	if len(this.noMethod) > 0 {
		for i := 0; i < tl; i++ {
			if ts[i].method == method {
				continue
			}

			var root = ts[i].root
			var value = root.getValue(path, c.params, false)
			if value.route != nil {
				c.handlers = this.allNoMethod
				this.handleError(c, http.StatusMethodNotAllowed, default405Body)
				return
			}
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
