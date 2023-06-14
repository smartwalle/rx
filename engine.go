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
}

func New() *Engine {
	var nEngine = &Engine{}
	nEngine.pool.New = func() interface{} {
		return &Context{}
	}
	return nEngine
}

func (this *Engine) Use(middleware ...HandlerFunc) {
	this.handlers = append(this.handlers, middleware...)
}

func (this *Engine) Load(provider RouteProvider) {
	this.provider = provider
}

func (this *Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	c := this.pool.Get().(*Context)
	c.reset(writer, request, this.handlers)

	this.handleHTTPRequest(c)

	this.pool.Put(c)
}

func (this *Engine) handleHTTPRequest(c *Context) {
	var route, err = this.provider.Match(c.Request)
	if err != nil {
		serveError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if route != nil {
		c.Route = route
		c.Next()

		if !c.IsAborted() {
			var target, err = c.Route.pick(c.Request)
			if err != nil {
				serveError(c, http.StatusInternalServerError, err.Error())
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

	serveError(c, http.StatusBadGateway, http.StatusText(http.StatusBadGateway))
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
