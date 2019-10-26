package rx

import (
	"fmt"
	"net/http"
	"strings"
)

type Router interface {
	Use(handlers ...HandlerFunc) Router

	Group(path string, handlers ...HandlerFunc) *RouterGroup

	GET(path string, handlers ...HandlerFunc)

	HEAD(path string, handlers ...HandlerFunc)

	POST(path string, handlers ...HandlerFunc)

	PUT(path string, handlers ...HandlerFunc)

	PATCH(path string, handlers ...HandlerFunc)

	DELETE(path string, handlers ...HandlerFunc)

	OPTIONS(path string, handlers ...HandlerFunc)
}

type RouterGroup struct {
	engine   *Engine
	trees    map[string]*methodTree
	basePath string
	handlers []HandlerFunc
	isRoot   bool
}

func newRouterGroup() *RouterGroup {
	var r = &RouterGroup{}
	r.trees = make(map[string]*methodTree)
	r.basePath = "/"
	return r
}

func (this *RouterGroup) Print() {
	for _, t := range this.trees {
		t.Print()
	}
}

func (this *RouterGroup) find(method, path string, isRegex bool) []*treeNode {
	path = cleanPath(path)

	var tree = this.trees[method]
	if tree == nil {
		return nil
	}

	return tree.find(path, isRegex)
}

func (this *RouterGroup) Use(handlers ...HandlerFunc) Router {
	this.handlers = append(this.handlers, handlers...)
	return this.returnObj()
}

func (this *RouterGroup) Group(path string, handlers ...HandlerFunc) *RouterGroup {
	var r = newRouterGroup()
	r.engine = this.engine
	r.trees = this.trees
	r.basePath = cleanPath(joinPaths(this.basePath, path))
	r.handlers = this.combineHandlers(handlers)
	return r
}

func (this *RouterGroup) Break(method, path string) {
	var tree = this.trees[method]
	if tree != nil {
		asset(path[0] == '/', "path must begin with '/'")
		path = cleanPath(path)
		tree.clean(path)
	}
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

func (this *RouterGroup) handle(method, path string, handlers HandlerChain) {
	path = cleanPath(joinPaths(this.basePath, path))

	asset(method != "", "HTTP method can not be empty")
	asset(path[0] == '/', "path must begin with '/'")
	asset(len(handlers) > 0, "there must be at least one handler")

	method = strings.ToUpper(method)

	var nHandlers = this.combineHandlers(handlers)

	var tree = this.trees[method]
	if tree == nil {
		tree = newMethodTree(method)
		this.trees[method] = tree
	}
	tree.add(path, nHandlers)

	logger.Output(3, fmt.Sprintf("%-8s %-30s --> %s (%d handlers)\n", method, path, nameOfFunction(nHandlers.Last()), nHandlers.Len()))
}

func (this *RouterGroup) combineHandlers(handlers HandlerChain) HandlerChain {
	if len(this.handlers) == 0 && len(handlers) == 0 {
		return nil
	}

	var nHandlers = make(HandlerChain, len(this.handlers)+len(handlers))
	if len(this.handlers) > 0 {
		copy(nHandlers, this.handlers)
	}
	if len(handlers) > 0 {
		copy(nHandlers[len(this.handlers):], handlers)
	}
	return nHandlers
}

func (this *RouterGroup) returnObj() Router {
	if this.isRoot {
		return this.engine
	}
	return this
}
