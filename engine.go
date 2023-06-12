package rx

import (
	"github.com/smartwalle/rx/balancer"
	"github.com/smartwalle/rx/balancer/roundrobin"
	"net/http"
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

func (this *Engine) Add(path string, targets []string, opts ...Option) {
	var nTargets = make([]*url.URL, 0, len(targets))
	for _, target := range targets {
		var nURL, err = url.Parse(target)
		if err != nil {
			panic(err.Error())
		}
		nTargets = append(nTargets, nURL)
	}

	nRegexp, err := regexp.Compile(path)
	if err != nil {
		panic(err.Error())
	}

	var location = &Location{}
	location.Path = path
	location.handlers = this.combineHandlers(nil)
	location.regexp = nRegexp
	location.targets = nTargets

	for _, opt := range opts {
		if opt != nil {
			opt(this, location)
		}
	}

	if location.balancer == nil {
		nBalancer, err := this.getBalancer("").Build(nTargets)
		if err != nil {
			panic(err.Error())
		}
		location.balancer = nBalancer
	}

	this.locations = append(this.locations, location)
}
