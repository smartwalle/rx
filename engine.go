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

func (e *Engine) Use(middleware ...HandlerFunc) {
	e.handlers = append(e.handlers, middleware...)
}

func (e *Engine) NoRoute(handlers ...HandlerFunc) {
	e.noRoute.handlers = handlers
}

func (e *Engine) NoProxy(handlers ...HandlerFunc) {
	e.noProxy.handlers = handlers
}

func (e *Engine) HandleError(handler ErrorHandler) {
	if handler == nil {
		handler = defaultErrorHandler
	}
	e.error = handler
}

func (e *Engine) Load(provider RouteProvider) {
	if provider == nil {
		provider = &nilProvider{}
	}
	e.provider = provider
}

func (e *Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	c := e.pool.Get().(*Context)
	c.mWriter.reset(writer)
	c.Request = request
	c.reset()
	c.handlers = e.handlers
	c.error = e.error

	e.handleHTTPRequest(c)

	e.pool.Put(c)
}

func (e *Engine) handleHTTPRequest(c *Context) {
	route, err := e.provider.Match(c.Request)
	if err != nil {
		c.route = e.noRoute
		e.serveError(c, http.StatusBadGateway, err.Error())
		return
	}

	if route == nil {
		c.route = e.noRoute
		e.serveError(c, http.StatusBadGateway, http.StatusText(http.StatusBadGateway))
		return
	}

	pResult, err := route.pick(c.Request)
	if err != nil {
		c.route = e.noProxy
		e.serveError(c, http.StatusBadGateway, err.Error())
		return
	}

	if pResult.Proxy == nil {
		c.route = e.noProxy
		e.serveError(c, http.StatusBadGateway, http.StatusText(http.StatusBadGateway))
		return
	}

	c.proxy = pResult.Proxy
	c.target = pResult.Target
	c.route = route
	c.Next()
	c.mWriter.WriteHeaderNow()
}

func (e *Engine) serveError(c *Context, code int, message string) {
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
