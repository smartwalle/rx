package rx

import (
	"net/http"
	"strings"
)

type Router interface {
	Use(handlers ...HandlerFunc) Router

	Group(path string, handlers ...HandlerFunc) *RouterGroup

	Break(method, path string)

	Exist(method, path string) bool

	GET(path string, targets []string, handlers ...HandlerFunc)

	HEAD(path string, targets []string, handlers ...HandlerFunc)

	POST(path string, targets []string, handlers ...HandlerFunc)

	PUT(path string, targets []string, handlers ...HandlerFunc)

	PATCH(path string, targets []string, handlers ...HandlerFunc)

	DELETE(path string, targets []string, handlers ...HandlerFunc)

	OPTIONS(path string, targets []string, handlers ...HandlerFunc)

	Any(path string, targets []string, handlers ...HandlerFunc)
}

type RouterGroup struct {
	engine   *Engine
	basePath string
	handlers []HandlerFunc
}

func newRouterGroup() *RouterGroup {
	var r = &RouterGroup{}
	r.basePath = "/"
	return r
}

func (this *RouterGroup) Use(handlers ...HandlerFunc) Router {
	this.handlers = append(this.handlers, handlers...)
	return this
}

func (this *RouterGroup) Group(path string, handlers ...HandlerFunc) *RouterGroup {
	var r = newRouterGroup()
	r.engine = this.engine
	r.basePath = CleanPath(joinPaths(this.basePath, path))
	r.handlers = this.combineHandlers(handlers)
	return r
}

func (this *RouterGroup) Break(method, path string) {
	method = strings.ToUpper(method)
	path = CleanPath(path)
	this.engine.breakRoute(method, path)
}

func (this *RouterGroup) Exist(method, path string) bool {
	method = strings.ToUpper(method)
	path = CleanPath(path)
	return this.engine.existRoute(method, path)
}

func (this *RouterGroup) GET(path string, targets []string, handlers ...HandlerFunc) {
	this.handle(http.MethodGet, path, targets, handlers)
}

func (this *RouterGroup) HEAD(path string, targets []string, handlers ...HandlerFunc) {
	this.handle(http.MethodHead, path, targets, handlers)
}

func (this *RouterGroup) POST(path string, targets []string, handlers ...HandlerFunc) {
	this.handle(http.MethodPost, path, targets, handlers)
}

func (this *RouterGroup) PUT(path string, targets []string, handlers ...HandlerFunc) {
	this.handle(http.MethodPut, path, targets, handlers)
}

func (this *RouterGroup) PATCH(path string, targets []string, handlers ...HandlerFunc) {
	this.handle(http.MethodPatch, path, targets, handlers)
}

func (this *RouterGroup) DELETE(path string, targets []string, handlers ...HandlerFunc) {
	this.handle(http.MethodDelete, path, targets, handlers)
}

func (this *RouterGroup) OPTIONS(path string, targets []string, handlers ...HandlerFunc) {
	this.handle(http.MethodOptions, path, targets, handlers)
}

func (this *RouterGroup) Any(path string, targets []string, handlers ...HandlerFunc) {
	this.handle(http.MethodGet, path, targets, handlers)
	this.handle(http.MethodHead, path, targets, handlers)
	this.handle(http.MethodPost, path, targets, handlers)
	this.handle(http.MethodPut, path, targets, handlers)
	this.handle(http.MethodPatch, path, targets, handlers)
	this.handle(http.MethodDelete, path, targets, handlers)
	this.handle(http.MethodConnect, path, targets, handlers)
	this.handle(http.MethodOptions, path, targets, handlers)
	this.handle(http.MethodTrace, path, targets, handlers)
}

func (this *RouterGroup) handle(method, path string, targets []string, handlers HandlersChain) {
	method = strings.ToUpper(method)
	path = CleanPath(joinPaths(this.basePath, path))
	var nHandlers = this.combineHandlers(handlers)
	this.engine.addRoute(method, path, targets, nHandlers)
}

func (this *RouterGroup) combineHandlers(handlers HandlersChain) HandlersChain {
	var hLen1 = len(this.handlers)
	var hLen2 = len(handlers)
	if hLen1 == 0 && hLen2 == 0 {
		return nil
	}

	var nHandlers = make(HandlersChain, len(this.handlers)+len(handlers))
	if hLen1 > 0 {
		copy(nHandlers, this.handlers)
	}
	if hLen2 > 0 {
		copy(nHandlers[hLen1:], handlers)
	}
	return nHandlers
}
