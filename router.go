package rx

import (
	"net/http"
	"strings"
)

type Router interface {
	Use(handlers ...HandlerFunc) Router

	Group(path string, handlers ...HandlerFunc) *RouterGroup

	Break(method, path string)

	GET(path string, handlers ...HandlerFunc)

	HEAD(path string, handlers ...HandlerFunc)

	POST(path string, handlers ...HandlerFunc)

	PUT(path string, handlers ...HandlerFunc)

	PATCH(path string, handlers ...HandlerFunc)

	DELETE(path string, handlers ...HandlerFunc)

	OPTIONS(path string, handlers ...HandlerFunc)

	Any(path string, handlers ...HandlerFunc)
}

type RouterGroup struct {
	engine   *Engine
	trees    methodTrees
	basePath string
	handlers []HandlerFunc
	isRoot   bool
}

func newRouterGroup() *RouterGroup {
	var r = &RouterGroup{}
	r.trees = make(methodTrees, 0, 8)
	r.basePath = "/"
	return r
}

func (this *RouterGroup) Use(handlers ...HandlerFunc) Router {
	this.handlers = append(this.handlers, handlers...)
	return this.returnObj()
}

func (this *RouterGroup) Group(path string, handlers ...HandlerFunc) *RouterGroup {
	var r = newRouterGroup()
	r.engine = this.engine
	r.trees = this.trees
	r.basePath = CleanPath(joinPaths(this.basePath, path))
	r.handlers = this.combineHandlers(handlers)
	return r
}

func (this *RouterGroup) Break(method, path string) {
	method = strings.ToUpper(method)
	path = CleanPath(path)
	this.engine.breakRoute(method, path)
}

func (this *RouterGroup) GET(path string, handlers ...HandlerFunc) {
	this.handle(http.MethodGet, path, handlers)
}

func (this *RouterGroup) HEAD(path string, handlers ...HandlerFunc) {
	this.handle(http.MethodHead, path, handlers)
}

func (this *RouterGroup) POST(path string, handlers ...HandlerFunc) {
	this.handle(http.MethodPost, path, handlers)
}

func (this *RouterGroup) PUT(path string, handlers ...HandlerFunc) {
	this.handle(http.MethodPut, path, handlers)
}

func (this *RouterGroup) PATCH(path string, handlers ...HandlerFunc) {
	this.handle(http.MethodPatch, path, handlers)
}

func (this *RouterGroup) DELETE(path string, handlers ...HandlerFunc) {
	this.handle(http.MethodDelete, path, handlers)
}

func (this *RouterGroup) OPTIONS(path string, handlers ...HandlerFunc) {
	this.handle(http.MethodOptions, path, handlers)
}

func (this *RouterGroup) Any(path string, handlers ...HandlerFunc) {
	this.handle(http.MethodGet, path, handlers)
	this.handle(http.MethodHead, path, handlers)
	this.handle(http.MethodPost, path, handlers)
	this.handle(http.MethodPut, path, handlers)
	this.handle(http.MethodPatch, path, handlers)
	this.handle(http.MethodDelete, path, handlers)
	this.handle(http.MethodConnect, path, handlers)
	this.handle(http.MethodOptions, path, handlers)
	this.handle(http.MethodTrace, path, handlers)
}

func (this *RouterGroup) handle(method, path string, handlers HandlerChain) {
	method = strings.ToUpper(method)
	path = CleanPath(joinPaths(this.basePath, path))
	var nHandlers = this.combineHandlers(handlers)
	this.engine.addRoute(method, path, nHandlers)
}

func (this *RouterGroup) combineHandlers(handlers HandlerChain) HandlerChain {
	var hLen1 = len(this.handlers)
	var hLen2 = len(handlers)
	if hLen1 == 0 && hLen2 == 0 {
		return nil
	}

	var nHandlers = make(HandlerChain, len(this.handlers)+len(handlers))
	if hLen1 > 0 {
		copy(nHandlers, this.handlers)
	}
	if hLen2 > 0 {
		copy(nHandlers[hLen1:], handlers)
	}
	return nHandlers
}

func (this *RouterGroup) returnObj() Router {
	if this.isRoot {
		return this.engine
	}
	return this
}
