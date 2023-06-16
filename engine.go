package rx

import (
	"log"
	"net/http"
	"sync"
)

type HandlerFunc func(c *Context)

type HandlersChain []HandlerFunc

type ErrorHandler func(c *Context, err error)

type Engine struct {
	pool     sync.Pool
	handlers HandlersChain
	provider RouteProvider

	noRoute *Route
	noProxy *Route
	error   ErrorHandler
}

func New() *Engine {
	var nEngine = &Engine{}
	nEngine.pool.New = func() interface{} {
		return &Context{}
	}
	nEngine.provider = &nilProvider{}
	nEngine.noRoute = &Route{}
	nEngine.noProxy = &Route{}
	nEngine.error = defaultErrorHandler
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

func (this *Engine) HandleError(handler ErrorHandler) {
	if handler == nil {
		handler = defaultErrorHandler
	}
	this.error = handler
}

func (this *Engine) Load(provider RouteProvider) {
	if provider == nil {
		provider = &nilProvider{}
	}
	this.provider = provider
}

func (this *Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	c := this.pool.Get().(*Context)
	c.mWriter.reset(writer)
	c.Request = request
	c.reset()
	c.handlers = this.handlers
	c.error = this.error

	this.handleHTTPRequest(c)

	this.pool.Put(c)
}

func (this *Engine) handleHTTPRequest(c *Context) {
	route, err := this.provider.Match(c.Request)
	if err != nil {
		c.route = this.noRoute
		this.serveError(c, http.StatusBadGateway, err.Error())
		return
	}

	if route == nil {
		c.route = this.noRoute
		this.serveError(c, http.StatusBadGateway, http.StatusText(http.StatusBadGateway))
		return
	}

	pResult, err := route.pick(c.Request)
	if err != nil {
		c.route = this.noProxy
		this.serveError(c, http.StatusBadGateway, err.Error())
		return
	}

	if pResult.Proxy == nil {
		c.route = this.noProxy
		this.serveError(c, http.StatusBadGateway, http.StatusText(http.StatusBadGateway))
		return
	}

	c.proxy = pResult.Proxy
	c.target = pResult.Target
	c.route = route
	c.Next()
	c.mWriter.WriteHeaderNow()
}

func (this *Engine) serveError(c *Context, code int, message string) {
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

func defaultErrorHandler(c *Context, err error) {
	log.Printf("proxy error: %v", err)
	c.AbortWithStatus(http.StatusInternalServerError)
}
