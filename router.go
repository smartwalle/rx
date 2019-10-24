package rx

import (
	"fmt"
	"net/http"
	"strings"
)

type Router struct {
	trees map[string]*Tree
}

func NewRouter() *Router {
	var r = &Router{}
	r.trees = make(map[string]*Tree)
	return r
}

func (this *Router) Print() {
	for _, t := range this.trees {
		t.Print()
	}
}

func (this *Router) find(method, path string, isRegex bool) []*Node {
	path = cleanPath(path)

	var tree = this.trees[method]
	if tree == nil {
		return nil
	}

	return tree.Find(path, isRegex)
}

func (this *Router) findOne(method, path string) *Node {
	path = cleanPath(path)

	var tree = this.trees[method]
	if tree == nil {
		return nil
	}

	return tree.FindOne(path)
}

func (this *Router) GET(path string, handlers ...http.HandlerFunc) {
	this.Handle(http.MethodGet, path, handlers...)
}

func (this *Router) Handle(method, path string, handlers ...http.HandlerFunc) {
	path = cleanPath(path)
	method = strings.ToUpper(method)

	asset(method != "", "HTTP method can not be empty")
	asset(path[0] == '/', "path must begin with '/'")
	asset(len(handlers) > 0, "there must be at least one handler")

	var tree = this.trees[method]
	if tree == nil {
		tree = NewTree()
		this.trees[method] = tree
	}
	tree.Add(path, handlers...)
}

func (this *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var path = cleanPath(req.URL.Path)
	fmt.Println(path)

	var tree = this.trees[req.Method]
	if tree == nil {
		// TODO method not allowed
		return
	}

	var nodes = tree.Find(path, false)
	if len(nodes) > 0 {
		var node = nodes[0]
		if node.path == path && len(node.handlers) > 0 {
			for _, handler := range node.handlers {
				handler(w, req)
			}
		}
	} else {
		// TODO regex
	}

	// TODO bad request
}
