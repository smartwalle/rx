package rx

import (
	"github.com/smartwalle/rx/balancer"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
)

type Option func(engine *Engine, location *Location)

func WithHandler(handlers ...HandlerFunc) Option {
	return func(engine *Engine, location *Location) {
		if len(handlers) > 0 {
			location.handlers = engine.combineHandlers(handlers)
		}
	}
}

func WithBalancer(builder balancer.Builder) Option {
	return func(engine *Engine, location *Location) {
		if builder != nil {
			var err error
			location.balancer, err = builder.Build(location.targets)
			if err != nil {
				panic(err)
			}
		}
	}
}

type Location struct {
	Path     string
	handlers HandlersChain
	regexp   *regexp.Regexp
	targets  []*url.URL
	balancer balancer.Balancer
}

func (this *Location) Match(path string) bool {
	return this.regexp.MatchString(path)
}

func (this *Location) Pick(req *http.Request) (*httputil.ReverseProxy, error) {
	return this.balancer.Pick(req)
}
