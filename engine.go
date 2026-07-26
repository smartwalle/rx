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
	pool   sync.Pool
	router Router

	noRoute Route
	error   ErrorHandler
}

func New() *Engine {
	var nEngine = &Engine{}
	nEngine.pool.New = func() interface{} {
		return &Context{}
	}
	nEngine.router = &nonRouter{}
	nEngine.noRoute = NewRoute("")
	nEngine.error = defaultErrorHandler
	return nEngine
}

func (e *Engine) NoRoute(handlers ...HandlerFunc) {
	e.noRoute = NewRoute("", handlers...)
}

func (e *Engine) HandleError(handler ErrorHandler) {
	if handler == nil {
		handler = defaultErrorHandler
	}
	e.error = handler
}

func (e *Engine) UseRouter(router Router) {
	if router == nil {
		router = &nonRouter{}
	}
	e.router = router
}

func (e *Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	c := e.pool.Get().(*Context)
	c.Context = request.Context()
	c.Request = request

	c.reset(writer)
	c.error = e.error

	e.handleHTTPRequest(c)

	e.pool.Put(c)
}

func (e *Engine) handleHTTPRequest(c *Context) {
	route, params, err := e.router.Match(c.Request.Method, c.Request.URL.Path, c.params[:0])
	if err != nil {
		e.serveError(c, e.noRoute, http.StatusBadGateway, err.Error())
		return
	}

	if route == nil {
		e.serveError(c, e.noRoute, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}

	c.route = route
	c.params = params
	c.Next()
	c.writer.WriteHeaderNow()
}

func (e *Engine) serveError(c *Context, route Route, code int, message string) {
	c.writer.status = code
	c.route = route
	c.index = -1
	c.Next()

	if c.writer.Written() {
		return
	}

	if c.writer.Status() == code {
		c.Writer.WriteString(message)
		return
	}
	c.writer.WriteHeaderNow()
}

func defaultErrorHandler(c *Context, err error) {
	log.Printf("proxy error: %v", err)
	c.AbortWithStatus(http.StatusInternalServerError)
}
