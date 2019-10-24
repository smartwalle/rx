package rx

import (
	"net/http"
	"strings"
)

type Router interface {
	Use(handlers ...http.HandlerFunc) Router

	Group(path string, handlers ...http.HandlerFunc) Router

	GET(path string, handlers ...http.HandlerFunc)
}

type RouterGroup struct {
	trees    map[string]*Tree
	basePath string
	handlers []http.HandlerFunc
	engine   *Engine
	isRoot   bool
}

func newRouterGroup() *RouterGroup {
	var r = &RouterGroup{}
	r.trees = make(map[string]*Tree)
	r.basePath = "/"
	return r
}

func (this *RouterGroup) Print() {
	for _, t := range this.trees {
		t.Print()
	}
}

func (this *RouterGroup) find(method, path string, isRegex bool) []*Node {
	path = cleanPath(path)

	var tree = this.trees[method]
	if tree == nil {
		return nil
	}

	return tree.Find(path, isRegex)
}

func (this *RouterGroup) findOne(method, path string) *Node {
	path = cleanPath(path)

	var tree = this.trees[method]
	if tree == nil {
		return nil
	}

	return tree.FindOne(path)
}

func (this *RouterGroup) Use(handlers ...http.HandlerFunc) Router {
	this.handlers = append(this.handlers, handlers...)
	return this.returnObj()
}

func (this *RouterGroup) Group(path string, handlers ...http.HandlerFunc) Router {
	var r = newRouterGroup()
	r.trees = this.trees
	r.basePath = cleanPath(joinPaths(this.basePath, path))
	r.handlers = this.combineHandlers(handlers)
	return r
}

func (this *RouterGroup) GET(path string, handlers ...http.HandlerFunc) {
	this.Handle(http.MethodGet, path, handlers...)
}

func (this *RouterGroup) Handle(method, path string, handlers ...http.HandlerFunc) {
	path = cleanPath(joinPaths(this.basePath, path))

	asset(method != "", "HTTP method can not be empty")
	asset(path[0] == '/', "path must begin with '/'")
	asset(len(handlers) > 0, "there must be at least one handler")

	method = strings.ToUpper(method)

	var nHandlers = this.combineHandlers(handlers)

	var tree = this.trees[method]
	if tree == nil {
		tree = newTree()
		this.trees[method] = tree
	}
	tree.Add(path, nHandlers...)
}

func (this *RouterGroup) combineHandlers(handlers []http.HandlerFunc) []http.HandlerFunc {
	if len(this.handlers) == 0 && len(handlers) == 0 {
		return nil
	}

	var nHandlers = make([]http.HandlerFunc, len(this.handlers)+len(handlers))
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
