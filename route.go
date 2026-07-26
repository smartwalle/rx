package rx

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/smartwalle/rx/balancer"
	"github.com/smartwalle/rx/balancer/roundrobin"
)

var ErrProxyNotFound = errors.New("proxy not found")

type ProxyBuilder func(target *url.URL) *httputil.ReverseProxy

type Option func(opts *options)

func WithBalancer(name string) Option {
	return func(opts *options) {
		opts.balancer = name
	}
}

func WithProxyBuilder(builder ProxyBuilder) Option {
	return func(opts *options) {
		opts.builder = builder
	}
}

type options struct {
	balancer string
	builder  ProxyBuilder
}

func defaultProxyErrorHandler(writer http.ResponseWriter, request *http.Request, err error) {
	if wrapper, ok := writer.(writerWrapper); ok {
		wrapper.context.abortWithError(err)
	}
}

func (opts *options) buildBalancer(targets []*url.URL) (balancer.Balancer, error) {
	if opts.balancer == "" {
		opts.balancer = roundrobin.Name
	}

	var builder = GetBalancer(opts.balancer)
	if builder == nil {
		return nil, fmt.Errorf("unknown balancer %s", opts.balancer)
	}

	var proxies = make(map[*url.URL]*httputil.ReverseProxy)
	for _, target := range targets {
		var proxy = opts.buildProxy(target)
		if proxy.ErrorHandler == nil {
			proxy.ErrorHandler = defaultProxyErrorHandler
		}
		proxies[target] = proxy
	}

	var info = balancer.BuildInfo{
		Targets: proxies,
	}
	return builder.Build(info)
}

func (opts *options) buildProxy(target *url.URL) *httputil.ReverseProxy {
	if opts.builder != nil {
		return opts.builder(target)
	}
	return httputil.NewSingleHostReverseProxy(target)
}

// Route 表示 Router 已经匹配到的处理目标。
type Route interface {
	Path() string
	Handlers() HandlersChain
	Serve(c *Context) error
}

type proxyRoute struct {
	path     string
	handlers HandlersChain
	targets  []*url.URL
	balancer balancer.Balancer
}

// NewProxyRoute 创建反向代理路由。
func NewProxyRoute(path string, targets []string, handlers []HandlerFunc, opts ...Option) (Route, error) {
	var targetURLs = make([]*url.URL, 0, len(targets))
	for _, target := range targets {
		nURL, err := url.Parse(target)
		if err != nil {
			return nil, err
		}
		targetURLs = append(targetURLs, nURL)
	}

	var nOpts = &options{}
	for _, opt := range opts {
		if opt != nil {
			opt(nOpts)
		}
	}
	balancer, err := nOpts.buildBalancer(targetURLs)
	if err != nil {
		return nil, err
	}

	var route = &proxyRoute{}
	route.path = path
	route.handlers = handlers
	route.targets = targetURLs
	route.balancer = balancer

	return route, nil
}

func (route *proxyRoute) Path() string {
	return route.path
}

func (route *proxyRoute) Handlers() HandlersChain {
	return route.handlers
}

func (route *proxyRoute) Serve(c *Context) error {
	if route == nil || route.balancer == nil {
		return ErrProxyNotFound
	}
	result, err := route.balancer.Pick(c.Request)
	if err != nil {
		return err
	}
	if result.Proxy == nil {
		return ErrProxyNotFound
	}
	result.Proxy.ServeHTTP(writerWrapper{ResponseWriter: c.Writer, context: c}, c.Request)
	return nil
}

type localRoute struct {
	path     string
	handlers HandlersChain
}

// NewRoute 创建只执行本地处理函数的路由。
func NewRoute(path string, handlers ...HandlerFunc) Route {
	var route = &localRoute{}
	route.path = path
	route.handlers = handlers
	return route
}

func (route *localRoute) Path() string {
	return route.path
}

func (route *localRoute) Handlers() HandlersChain {
	return route.handlers
}

func (route *localRoute) Serve(_ *Context) error {
	return nil
}
