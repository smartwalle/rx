package rx

import (
	"net/http"
)

type RouteProvider interface {
	Match(req *http.Request) (*Route, error)
}

type nilProvider struct {
}

func (this *nilProvider) Match(req *http.Request) (*Route, error) {
	panic(`provider should be specified. 
var provider = rx.NewListProvider()
var engine = rx.New()
engine.Load(provider)`)
	return nil, nil
}

// ListProvider 路由信息管理
//
// 内部维护了一个 Route 对象的切片，可用于路由规则较少的场景。
type ListProvider struct {
	routes []*Route
}

func NewListProvider() *ListProvider {
	return &ListProvider{}
}

// Match 匹配路由
func (this *ListProvider) Match(req *http.Request) (*Route, error) {
	var path = req.URL.Path
	for _, route := range this.routes {
		if route.Match(path) {
			return route, nil
		}
	}
	return nil, nil
}

// Add 添加路由规则
//
// 由于 ListProvider 的 Match 方法是依次对路由进行匹配，所以规则 /user 应该先于规则 /u 添加。
//
// 如：
//
// provider.Add("/user", []string{"http://xxx"})
//
// provider.Add("/u", []string{"http://xxx"})
func (this *ListProvider) Add(pattern string, targets []string, opts ...Option) error {
	var route, err = NewRoute(pattern, targets, opts...)
	if err != nil {
		return err
	}
	this.routes = append(this.routes, route)
	return nil
}
