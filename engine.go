package rx

import (
	"fmt"
	"net/http"
	"sync"
)

type HandlerFunc func(c *Context)

type HandlersChain []HandlerFunc

type ErrorHandler func(c *Context, err error)

type Engine struct {
	pool     sync.Pool
	handlers HandlersChain
	router   Router

	noRoute Route
	error   ErrorHandler
}

func New() *Engine {
	var engine = &Engine{}
	engine.pool.New = func() any {
		return &Context{}
	}
	engine.router = &noop{}
	engine.noRoute = NewRoute("")
	engine.error = defaultErrorHandler
	return engine
}

func (e *Engine) Use(middleware ...HandlerFunc) {
	if len(middleware) == 0 {
		return
	}
	e.handlers = append(e.handlers, middleware...)
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
		router = &noop{}
	}
	e.router = router
}

func (e *Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	c := e.pool.Get().(*Context)
	c.Request = request

	c.reset(writer)
	c.error = e.error
	c.handlers = e.handlers

	e.handleHTTPRequest(c)

	e.pool.Put(c)
}

func (e *Engine) handleHTTPRequest(c *Context) {
	route, params, err := e.router.Match(c.Request.Method, c.Request.URL.Path, c.Params[:0])
	if err != nil {
		c.AbortWithError(err)
		return
	}

	if route == nil {
		e.serveError(c, e.noRoute, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}

	c.route = route
	c.Params = params
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
	c.AbortWithStatus(http.StatusInternalServerError)
	fmt.Fprintln(c.Writer, err.Error())
}
