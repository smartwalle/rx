package rx

import (
	"fmt"
	"github.com/smartwalle/rx/balancer/roundrobin"
	"net/http"
	"net/url"
	"sync"
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

	allNotFound HandlersChain
	notFound    HandlersChain

	allMethodNotAllowed HandlersChain
	methodNotAllowed    HandlersChain

	allBadGateway HandlersChain
	badGateway    HandlersChain

	allServiceUnavailable HandlersChain
	serviceUnavailable    HandlersChain
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
	this.rebuild502Handlers()
	this.rebuild503Handlers()
	return this
}

func (this *Engine) NotFound(handlers ...HandlerFunc) {
	this.notFound = handlers
	this.rebuild404Handlers()
}

func (this *Engine) MethodNotAllowed(handlers ...HandlerFunc) {
	this.methodNotAllowed = handlers
	this.rebuild405Handlers()
}

func (this *Engine) BadGateway(handlers ...HandlerFunc) {
	this.badGateway = handlers
	this.rebuild502Handlers()
}

func (this *Engine) ServiceUnavailable(handlers ...HandlerFunc) {
	this.serviceUnavailable = handlers
	this.rebuild503Handlers()
}

func (this *Engine) rebuild404Handlers() {
	this.allNotFound = this.combineHandlers(this.notFound)
}

func (this *Engine) rebuild405Handlers() {
	this.allMethodNotAllowed = this.combineHandlers(this.methodNotAllowed)
}

func (this *Engine) rebuild502Handlers() {
	this.allBadGateway = this.combineHandlers(this.badGateway)
}

func (this *Engine) rebuild503Handlers() {
	this.allServiceUnavailable = this.combineHandlers(this.serviceUnavailable)
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

	this.handleRequest(c)

	this.pool.Put(c)
}

func (this *Engine) handleRequest(c *Context) {
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
			var target, err = value.route.balancer.Pick(c.Request)
			// 502 错误
			if err != nil {
				c.handlers = this.allBadGateway
				this.handleError(c, http.StatusBadGateway, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadGateway), err.Error()))
				return
			}
			// 503 错误
			if target == nil {
				c.handlers = this.allServiceUnavailable
				this.handleError(c, http.StatusServiceUnavailable, http.StatusText(http.StatusServiceUnavailable))
				return
			}

			c.target = target
			c.handlers = value.route.handlers
			c.params = value.params

			c.exec()
			return
		}
	}

	// 405 错误
	if len(this.methodNotAllowed) > 0 {
		for i := 0; i < tl; i++ {
			if ts[i].method == method {
				continue
			}

			var root = ts[i].root
			var value = root.getValue(path, c.params, false)
			if value.route != nil {
				c.handlers = this.allMethodNotAllowed
				this.handleError(c, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
				return
			}
		}
	}

	// 404 错误
	c.handlers = this.allNotFound
	this.handleError(c, http.StatusNotFound, http.StatusText(http.StatusNotFound))
}

func (this *Engine) handleError(c *Context, status int, body string) {
	var w = c.Writer
	w.WriteHeader(status)

	c.Next()

	if w.Written() {
		return
	}

	w.WriteString(body)
}
