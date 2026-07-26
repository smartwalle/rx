package router

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/smartwalle/rx"
)

type methodTree struct {
	method string
	tree   *tree
}

type methodTrees []methodTree

func (trees methodTrees) get(method string) *tree {
	for _, t := range trees {
		if t.method == method {
			return t.tree
		}
	}
	return nil
}

// Router 提供 method/path 路由注册、分组和匹配能力。
type Router interface {
	rx.Router

	Use(handlers ...rx.HandlerFunc)
	Group(path string, handlers ...rx.HandlerFunc) Router
	Proxy(method string, path string, targets []string, opts ...rx.Option) error

	GET(path string, handlers ...rx.HandlerFunc) error
	POST(path string, handlers ...rx.HandlerFunc) error
	PUT(path string, handlers ...rx.HandlerFunc) error
	PATCH(path string, handlers ...rx.HandlerFunc) error
	DELETE(path string, handlers ...rx.HandlerFunc) error
	HEAD(path string, handlers ...rx.HandlerFunc) error
	OPTIONS(path string, handlers ...rx.HandlerFunc) error
}

type router struct {
	group
	trees methodTrees
}

var _ Router = (*router)(nil)
var _ Router = (*group)(nil)

// New 创建一个支持分组和 method 分发的 Router。
func New() Router {
	r := &router{}
	r.group = group{
		router: r,
	}
	return r
}

func (r *router) add(method string, route rx.Route) error {
	if r == nil {
		return nil
	}
	if method == "" {
		return fmt.Errorf("method must not be empty")
	}
	var t = r.getOrCreate(method)
	return t.add(route)
}

// Match 根据 method 和 path 匹配 Route。
func (r *router) Match(method string, path string, params rx.Params) (rx.Route, rx.Params, error) {
	if r == nil {
		return nil, params, nil
	}
	var t = r.trees.get(method)
	if t == nil {
		return nil, params, nil
	}
	return t.match(path, params)
}

func (r *router) getOrCreate(method string) *tree {
	if t := r.trees.get(method); t != nil {
		return t
	}
	var t = newTree()
	r.trees = append(r.trees, methodTree{
		method: method,
		tree:   t,
	})
	return t
}

type group struct {
	router   *router
	base     string
	handlers rx.HandlersChain
}

// Use 向当前分组追加中间件。
func (g *group) Use(handlers ...rx.HandlerFunc) {
	if g == nil || len(handlers) == 0 {
		return
	}
	g.handlers = combineHandlers(g.handlers, handlers)
}

// Group 创建子分组。
func (g *group) Group(path string, handlers ...rx.HandlerFunc) Router {
	if g == nil {
		return nil
	}
	return &group{
		router:   g.router,
		base:     joinPath(g.base, path),
		handlers: combineHandlers(g.handlers, handlers),
	}
}

func (g *group) Match(method string, path string, params rx.Params) (rx.Route, rx.Params, error) {
	if g == nil || g.router == nil {
		return nil, params, nil
	}
	return g.router.Match(method, path, params)
}

func (g *group) handle(method string, path string, handlers ...rx.HandlerFunc) error {
	if g == nil || g.router == nil {
		return nil
	}
	fullPath := joinPath(g.base, path)
	return g.router.add(method, rx.NewRoute(fullPath, combineHandlers(g.handlers, handlers)...))
}

// Proxy 注册指定 method 的反向代理路由。
func (g *group) Proxy(method string, path string, targets []string, opts ...rx.Option) error {
	if g == nil || g.router == nil {
		return nil
	}
	if !validMethod(method) {
		return fmt.Errorf("invalid method %q", method)
	}
	fullPath := joinPath(g.base, path)
	route, err := rx.NewProxyRoute(fullPath, targets, g.handlers, opts...)
	if err != nil {
		return err
	}
	return g.router.add(method, route)
}

func (g *group) GET(path string, handlers ...rx.HandlerFunc) error {
	return g.handle(http.MethodGet, path, handlers...)
}

func (g *group) POST(path string, handlers ...rx.HandlerFunc) error {
	return g.handle(http.MethodPost, path, handlers...)
}

func (g *group) PUT(path string, handlers ...rx.HandlerFunc) error {
	return g.handle(http.MethodPut, path, handlers...)
}

func (g *group) PATCH(path string, handlers ...rx.HandlerFunc) error {
	return g.handle(http.MethodPatch, path, handlers...)
}

func (g *group) DELETE(path string, handlers ...rx.HandlerFunc) error {
	return g.handle(http.MethodDelete, path, handlers...)
}

func (g *group) HEAD(path string, handlers ...rx.HandlerFunc) error {
	return g.handle(http.MethodHead, path, handlers...)
}

func (g *group) OPTIONS(path string, handlers ...rx.HandlerFunc) error {
	return g.handle(http.MethodOptions, path, handlers...)
}

func combineHandlers(base rx.HandlersChain, handlers []rx.HandlerFunc) rx.HandlersChain {
	if len(base) == 0 && len(handlers) == 0 {
		return nil
	}
	combined := make(rx.HandlersChain, 0, len(base)+len(handlers))
	combined = append(combined, base...)
	combined = append(combined, handlers...)
	return combined
}

func validMethod(method string) bool {
	switch method {
	case http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodHead,
		http.MethodOptions,
		http.MethodConnect,
		http.MethodTrace:
		return true
	default:
		return false
	}
}

func joinPath(base string, path string) string {
	if path == "" {
		if base == "" {
			return "/"
		}
		return base
	}
	if base == "" || base == "/" {
		if strings.HasPrefix(path, "/") {
			return path
		}
		return "/" + path
	}

	if strings.HasSuffix(base, "/") {
		if strings.HasPrefix(path, "/") {
			return base + path[1:]
		}
		return base + path
	}

	if strings.HasPrefix(path, "/") {
		return base + path
	}
	return base + "/" + path
}
