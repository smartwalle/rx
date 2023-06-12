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

func (this *Engine) GetBalancer(name string) balancer.Builder {
	if name == "" {
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
		if location.Regexp.MatchString(path) {
			c.Location = location
			c.handlers = this.handlers
			c.Next()

			if !c.IsAborted() {
				var target, err = c.Location.Balancer.Pick(c.Request)
				if err != nil {

				}
				target.ServeHTTP(c.Writer, c.Request)
			}
			c.mWriter.WriteHeaderNow()
			return
		}
	}

	c.AbortWithStatus(http.StatusNotFound)
	c.Writer.WriteString(http.StatusText(http.StatusNotFound))
}

func (this *Engine) Add(path string, targets []string) {
	var nTargets = make([]*url.URL, 0, len(targets))

	for _, target := range targets {
		var nURL, err = url.Parse(target)
		if err != nil {
			panic(err.Error())
		}
		nTargets = append(nTargets, nURL)
	}

	nBalancer, err := this.GetBalancer("").Build(nTargets)
	if err != nil {
		panic(err.Error())
	}
	nRegexp, err := regexp.Compile(path)
	if err != nil {
		panic(err.Error())
	}

	var location = &Location{}
	location.Path = path
	location.Regexp = nRegexp
	location.Targets = nTargets
	location.Balancer = nBalancer

	this.locations = append(this.locations, location)
}

type Location struct {
	Path     string
	Regexp   *regexp.Regexp
	Targets  []*url.URL
	Balancer balancer.Balancer
}
