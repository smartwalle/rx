package rx

import (
	"fmt"
	"github.com/smartwalle/rx/balancer"
	"github.com/smartwalle/rx/balancer/roundrobin"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
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

func (this *options) buildBalancer(targets []*url.URL) (balancer.Balancer, error) {
	if this.balancer == "" {
		this.balancer = roundrobin.Name
	}

	var bBuilder = GetBalancer(this.balancer)
	if bBuilder == nil {
		return nil, fmt.Errorf("unknown balancer %s", this.balancer)
	}

	var proxies = make(map[*url.URL]*httputil.ReverseProxy)
	for _, target := range targets {
		var proxy = this.buildProxy(target)

		if proxy != nil {
			proxies[target] = proxy
		}
	}

	var info = balancer.BuildInfo{
		Targets: proxies,
	}
	return bBuilder.Build(info)
}

func (this *options) buildProxy(target *url.URL) *httputil.ReverseProxy {
	if this.builder != nil {
		return this.builder(target)
	}
	return httputil.NewSingleHostReverseProxy(target)
}

type Route struct {
	pattern string
	regexp  *regexp.Regexp
	targets []*url.URL

	handlers HandlersChain
	balancer balancer.Balancer
}

func NewRoute(pattern string, targets []string, opts ...Option) (*Route, error) {
	nRegexp, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	var nTargets = make([]*url.URL, 0, len(targets))
	for _, target := range targets {
		nURL, err := url.Parse(target)
		if err != nil {
			return nil, err
		}
		nTargets = append(nTargets, nURL)
	}

	var nOpts = &options{}
	for _, opt := range opts {
		if opt != nil {
			opt(nOpts)
		}
	}
	nBalancer, err := nOpts.buildBalancer(nTargets)
	if err != nil {
		return nil, err
	}

	var route = &Route{}
	route.pattern = pattern
	route.regexp = nRegexp
	route.targets = nTargets
	route.handlers = nOpts.handlers
	route.balancer = nBalancer

	return route, nil
}

func (this *Route) Pattern() string {
	return this.pattern
}

func (this *Route) Match(path string) bool {
	return this.regexp.MatchString(path)
}

func (this *Route) pick(req *http.Request) (balancer.PickResult, error) {
	return this.balancer.Pick(req)
}
