package rx

import (
	"net/http"
	"sync"
)

type HandlerFunc func(c *Context)

type HandlersChain []HandlerFunc

type Engine struct {
	handlers HandlersChain
	provider RouteProvider
	pool     sync.Pool

	noRoute *Route
	noProxy *Route
}

func New() *Engine {
	var nEngine = &Engine{}
	nEngine.pool.New = func() interface{} {
		return &Context{}
	}
	nEngine.noRoute = &Route{}
	nEngine.noProxy = &Route{}
	return nEngine
}

func (this *Engine) Use(middleware ...HandlerFunc) {
	this.handlers = append(this.handlers, middleware...)
}

func (this *Engine) NoRoute(handlers ...HandlerFunc) {
	this.noRoute.handlers = handlers
}

func (this *Engine) NoProxy(handlers ...HandlerFunc) {
	this.noProxy.handlers = handlers
}

func (this *Engine) Load(provider RouteProvider) {
	this.provider = provider
}

func (this *Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	c := this.pool.Get().(*Context)
	c.mWriter.reset(writer)
	c.Request = request
	c.reset()
	c.handlers = this.handlers

	this.handleHTTPRequest(c)

	this.pool.Put(c)
}

func (this *Engine) handleHTTPRequest(c *Context) {
	route, err := this.provider.Match(c.Request)
	if err != nil || route == nil {
		c.Route = this.noRoute
		this.handleError(c, http.StatusBadGateway, http.StatusText(http.StatusBadGateway))
		return
	}

	proxy, err := route.pick(c.Request)
	if err != nil || proxy == nil {
		c.Route = this.noProxy
		this.handleError(c, http.StatusBadGateway, http.StatusText(http.StatusBadGateway))
		return
	}

	c.proxy = proxy
	c.Route = route
	c.Next()
	c.mWriter.WriteHeaderNow()
}

func (this *Engine) handleError(c *Context, code int, message string) {
	c.mWriter.status = code
	c.Next()

	if c.mWriter.Written() {
		return
	}

	if c.mWriter.Status() == code {
		//c.mWriter.Header()[kContentType] = kContentTypeText
		c.Writer.WriteString(message)
		return
	}
	c.mWriter.WriteHeaderNow()
}
