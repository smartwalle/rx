package rx

import "net/http"

type RouteProvider interface {
	Match(req *http.Request) (*Route, error)
}

type ListProvider struct {
	routes []*Route
}

func NewListProvider() *ListProvider {
	return &ListProvider{}
}

func (this *ListProvider) Match(req *http.Request) (*Route, error) {
	var path = req.URL.Path
	for _, route := range this.routes {
		if route.Match(path) {
			return route, nil
		}
	}
	return nil, nil
}

func (this *ListProvider) Add(pattern string, targets []string, opts ...Option) error {
	var route, err = NewRoute(pattern, targets, opts...)
	if err != nil {
		return err
	}
	this.routes = append(this.routes, route)
	return nil
}
