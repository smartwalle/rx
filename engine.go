package rx

import (
	"net/http"
	"sync"
)

type HandlerFunc func(c *Context)

type HandlersChain []HandlerFunc

type ErrorEncoder func(c *Context, err error)

type Engine struct {
	pool     sync.Pool
	handlers HandlersChain
	router   Router

	notFound     Route
	errorEncoder ErrorEncoder
}

func New() *Engine {
	var engine = &Engine{}
	engine.pool.New = func() any {
		return &Context{}
	}
	engine.router = &noop{}
	engine.notFound = NewRoute("")
	engine.errorEncoder = defaultErrorEncoder
	return engine
}

// Use 追加全局中间件。
//
// 全局中间件对所有请求生效，按注册顺序依次执行，并且先于路由自身的处理函数。
// 多次调用会依次累加；需要终止后续处理时调用 Context.Abort。
func (e *Engine) Use(middleware ...HandlerFunc) {
	if len(middleware) == 0 {
		return
	}
	e.handlers = append(e.handlers, middleware...)
}

// UseNotFoundHandler 设置路由未匹配时的兜底处理器。
//
// 该处理器只在没有任何路由匹配时执行（即返回 404），并且位于全局中间件之后。
// 若处理器没有输出任何内容，Engine 会输出默认的 404 提示。
// 多次调用以最后一次为准。
func (e *Engine) UseNotFoundHandler(handlers ...HandlerFunc) {
	e.notFound = NewRoute("", handlers...)
}

// UseErrorEncoder 设置错误编码器。
//
// 请求处理过程中发生错误时，ErrorEncoder 负责把错误转换成响应。
// 传入 nil 时恢复为默认实现，即返回 500 并输出错误信息。
// 多次调用以最后一次为准。
func (e *Engine) UseErrorEncoder(encoder ErrorEncoder) {
	if encoder == nil {
		encoder = defaultErrorEncoder
	}
	e.errorEncoder = encoder
}

// UseRouter 设置路由表。
//
// Router 负责把 method 和 path 匹配到具体的处理函数，是 Engine 处理请求的前提。
// 传入 nil 时使用默认实现，此时所有请求都会返回 400 并提示如何正确指定路由。
// 多次调用以最后一次为准。
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
	c.errorEncoder = e.errorEncoder
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
		e.serveNotFound(c, e.notFound, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}

	c.route = route
	c.Params = params
	c.Next()
	c.writer.WriteHeaderNow()
}

// serveNotFound 执行兜底处理器；若处理器没有输出任何内容，则输出默认提示。
func (e *Engine) serveNotFound(c *Context, route Route, code int, message string) {
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

func defaultErrorEncoder(c *Context, err error) {
	c.AbortWithStatus(http.StatusInternalServerError)
	c.Writer.WriteString(err.Error())
}
