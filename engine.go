package rx

import (
	"github.com/smartwalle/rx/balancer"
	"github.com/smartwalle/rx/balancer/roundrobin"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"sync"
)

type HandlerFunc func(c *Context)

type HandlersChain []HandlerFunc

type Engine struct {
	handlers HandlersChain

	balancers map[string]balancer.Builder
	locations []*Location

	pool sync.Pool

	ProxyBuilder func(target *url.URL) (*httputil.ReverseProxy, error)
}

func New() *Engine {
	var nEngine = &Engine{}
	nEngine.balancers = make(map[string]balancer.Builder)
	nEngine.pool.New = func() interface{} {
		return &Context{}
	}
	nEngine.RegisterBalancer(roundrobin.New())
	return nEngine
}

func (this *Engine) Use(middleware ...HandlerFunc) {
	this.handlers = append(this.handlers, middleware...)
}

func (this *Engine) RegisterBalancer(builder balancer.Builder) {
	if builder != nil && builder.Name() != "" {
		this.balancers[builder.Name()] = builder
	}
}

func (this *Engine) getBalancer(name string) balancer.Builder {
	if name == "" || this.balancers[name] == nil {
		name = roundrobin.Name
	}
	return this.balancers[name]
}

func (this *Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	c := this.pool.Get().(*Context)
	c.reset(writer, request)

	this.handleHTTPRequest(c)

	this.pool.Put(c)
}

func (this *Engine) handleHTTPRequest(c *Context) {
	var path = c.Request.URL.Path
	for _, location := range this.locations {
		if location.Match(path) {
			c.Location = location
			c.Next()

			if !c.IsAborted() {
				var target, err = c.Location.Pick(c.Request)
				if err != nil {
					serveError(c, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
					return
				}
				if target == nil {
					serveError(c, http.StatusBadGateway, http.StatusText(http.StatusBadGateway))
					return
				}
				target.ServeHTTP(c.Writer, c.Request)
			}
			c.mWriter.WriteHeaderNow()
			return
		}
	}
	c.AbortWithStatus(http.StatusBadGateway)
	c.Writer.WriteString(http.StatusText(http.StatusBadGateway))
}

func serveError(c *Context, code int, message string) {
	c.mWriter.status = code
	c.Next()
	if c.mWriter.Written() {
		return
	}
	if c.mWriter.Status() == code {
		c.mWriter.Header()[kContentType] = kContentTypeText
		c.Writer.WriteString(message)
		return
	}
	c.mWriter.WriteHeaderNow()
}

func (this *Engine) combineHandlers(handlers HandlersChain) HandlersChain {
	finalSize := len(this.handlers) + len(handlers)
	mergedHandlers := make(HandlersChain, finalSize)
	copy(mergedHandlers, this.handlers)
	copy(mergedHandlers[len(this.handlers):], handlers)
	return mergedHandlers
}

func (this *Engine) Add(path string, targets []string, opts ...Option) error {
	var location, err = this.BuildLocation(path, targets, opts...)
	if err != nil {
		return err
	}
	this.locations = append(this.locations, location)
	return nil
}

func (this *Engine) UpdateLocations(locations []*Location) {
	this.locations = locations
}

func (this *Engine) BuildLocation(path string, targets []string, opts ...Option) (*Location, error) {
	var nTargets = make([]*url.URL, 0, len(targets))
	for _, target := range targets {
		nURL, err := url.Parse(target)
		if err != nil {
			return nil, err
		}
		nTargets = append(nTargets, nURL)
	}

	nRegexp, err := regexp.Compile(path)
	if err != nil {
		return nil, err
	}

	var location = &Location{}
	location.Path = path
	location.handlers = this.combineHandlers(nil)
	location.regexp = nRegexp
	location.targets = nTargets

	for _, opt := range opts {
		if opt != nil {
			if err = opt(this, location); err != nil {
				return nil, err
			}
		}
	}

	if location.balancer == nil {
		info, nErr := this.buildBalancerBuildInfo(location.targets)
		if err != nil {
			return nil, err
		}

		nBalancer, nErr := this.getBalancer("").Build(info)
		if nErr != nil {
			return nil, nErr
		}
		location.balancer = nBalancer
	}
	return location, nil
}

func (this *Engine) buildBalancerBuildInfo(targets []*url.URL) (balancer.BuildInfo, error) {
	var builder = this.ProxyBuilder
	if builder == nil {
		builder = this.defaultReverseProxyBuilder
	}

	var proxies = make(map[*url.URL]*httputil.ReverseProxy)
	for _, target := range targets {
		var proxy, err = builder(target)
		if err != nil {
			return balancer.BuildInfo{}, err
		}
		if proxy != nil {
			proxies[target] = proxy
		}
	}
	return balancer.BuildInfo{Targets: proxies}, nil
}

func (this *Engine) defaultReverseProxyBuilder(target *url.URL) (*httputil.ReverseProxy, error) {
	return httputil.NewSingleHostReverseProxy(target), nil
}
