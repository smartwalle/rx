package rx

import (
	"net/http"
	"strings"
)

type Router interface {
	Use(handlers ...HandlerFunc) Router

	Group(path string, handlers ...HandlerFunc) Router

	GET(path string, handlers ...HandlerFunc)
}

type RouterGroup struct {
	trees    map[string]*methodTree
	basePath string
	handlers []HandlerFunc
	engine   *Engine
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

func (this *RouterGroup) find(method, path string, isRegex bool) []*pathNode {
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

func (this *RouterGroup) Group(path string, handlers ...HandlerFunc) Router {
	var r = newRouterGroup()
	r.engine = this.engine
	r.trees = this.trees
	r.basePath = cleanPath(joinPaths(this.basePath, path))
	r.handlers = this.combineHandlers(handlers)
	return r
}

func (this *RouterGroup) GET(path string, handlers ...HandlerFunc) {
	this.Handle(http.MethodGet, path, handlers...)
}

func (this *RouterGroup) Handle(method, path string, handlers ...HandlerFunc) {
	path = cleanPath(joinPaths(this.basePath, path))

	asset(method != "", "HTTP method can not be empty")
	asset(path[0] == '/', "path must begin with '/'")
	asset(len(handlers) > 0, "there must be at least one handler")

	method = strings.ToUpper(method)

	var nHandlers = this.combineHandlers(handlers)

	var tree = this.trees[method]
	if tree == nil {
		tree = newMethodTree()
		this.trees[method] = tree
	}
	tree.add(path, nHandlers...)
}

func (this *RouterGroup) combineHandlers(handlers []HandlerFunc) []HandlerFunc {
	if len(this.handlers) == 0 && len(handlers) == 0 {
		return nil
	}

	var nHandlers = make([]HandlerFunc, len(this.handlers)+len(handlers))
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
