package rx

import (
	"github.com/smartwalle/rx/balancer"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
)

type Option func(engine *Engine, location *Location) error

func WithHandler(handlers ...HandlerFunc) Option {
	return func(engine *Engine, location *Location) error {
		if len(handlers) > 0 {
			location.handlers = engine.combineHandlers(handlers)
		}
		return nil
	}
}

func WithBalancer(builder balancer.Builder) Option {
	return func(engine *Engine, location *Location) error {
		if builder != nil {
			info, nErr := engine.buildBalancerBuildInfo(location.targets)
			if nErr != nil {
				return nErr
			}
			nBalancer, nErr := builder.Build(info)
			if nErr != nil {
				return nErr
			}
			location.balancer = nBalancer
		}
		return nil
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

func (this *Location) pick(req *http.Request) (*httputil.ReverseProxy, error) {
	return this.balancer.Pick(req)
}
