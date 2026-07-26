package rx

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/smartwalle/rx/balancer"
	"github.com/smartwalle/rx/balancer/roundrobin"
)

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

func WithHandlers(handlers ...HandlerFunc) Option {
	return func(opts *options) {
		opts.handlers = handlers
	}
}

type options struct {
	balancer string
	builder  ProxyBuilder
	handlers HandlersChain
}

func defaultProxyErrorHandler(writer http.ResponseWriter, request *http.Request, err error) {
	if wrapper, ok := writer.(rWriterWrapper); ok {
		wrapper.context.abortWithError(err)
	}
}

func (opts *options) buildBalancer(targets []*url.URL) (balancer.Balancer, error) {
	if opts.balancer == "" {
		opts.balancer = roundrobin.Name
	}

	var bBuilder = GetBalancer(opts.balancer)
	if bBuilder == nil {
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
	return bBuilder.Build(info)
}

func (opts *options) buildProxy(target *url.URL) *httputil.ReverseProxy {
	if opts.builder != nil {
		return opts.builder(target)
	}
	return httputil.NewSingleHostReverseProxy(target)
}

// Route 表示 Router 已经匹配到的处理目标。
type Route interface {
	Pattern() string
	Handlers() HandlersChain
	Serve(c *Context) error
}

type ProxyRoute struct {
	pattern  string
	handlers HandlersChain
	targets  []*url.URL
	balancer balancer.Balancer
}

// NewProxyRoute 创建反向代理路由。
func NewProxyRoute(pattern string, targets []string, opts ...Option) (*ProxyRoute, error) {
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

	var route = &ProxyRoute{}
	route.pattern = pattern
	route.handlers = nOpts.handlers
	route.targets = targetURLs
	route.balancer = balancer

	return route, nil
}

func (route *ProxyRoute) Pattern() string {
	return route.pattern
}

func (route *ProxyRoute) Handlers() HandlersChain {
	return route.handlers
}

func (route *ProxyRoute) Serve(c *Context) error {
	if route == nil || route.balancer == nil {
		return fmt.Errorf("no proxy")
	}
	result, err := route.balancer.Pick(c.Request)
	if err != nil {
		return err
	}
	if result.Proxy == nil {
		return fmt.Errorf("no proxy")
	}
	result.Proxy.ServeHTTP(rWriterWrapper{ResponseWriter: c.Writer, context: c}, c.Request)
	return nil
}

type LocalRoute struct {
	pattern  string
	handlers HandlersChain
}

// NewLocalRoute 创建只执行本地处理函数的路由。
func NewLocalRoute(pattern string, handlers ...HandlerFunc) *LocalRoute {
	var route = &LocalRoute{}
	route.pattern = pattern
	route.handlers = handlers
	return route
}

func (route *LocalRoute) Pattern() string {
	return route.pattern
}

func (route *LocalRoute) Handlers() HandlersChain {
	return route.handlers
}

func (route *LocalRoute) Serve(c *Context) error {
	return nil
}
