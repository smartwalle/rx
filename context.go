package rx

import (
	"context"
	"net/http"
	"sync"
)

const (
	kContentType = "Content-Type"
)

// Param 是一次路径匹配得到的单个参数。
type Param struct {
	// Key 是规则中声明的参数名，不包含前缀的 : 或 *。
	Key string

	// Value 是输入路径中被该参数捕获到的原始字符串。
	Value string
}

// Params 是按规则中参数出现顺序保存的参数列表。
type Params []Param

// Get 返回指定名称的参数。
func (ps Params) Get(name string) (string, bool) {
	for _, item := range ps {
		if item.Key == name {
			return item.Value, true
		}
	}
	return "", false
}

// ByName 返回指定名称的参数；不存在时返回空字符串。
func (ps Params) ByName(name string) string {
	value, _ := ps.Get(name)
	return value
}

type Context struct {
	Request *http.Request
	Writer  ResponseWriter
	Params  Params

	writer responseWriter

	index   int
	aborted bool

	handlers     HandlersChain
	route        Route
	errorEncoder ErrorEncoder

	mu   sync.RWMutex
	Keys map[string]any
}

// UseContext 替换当前请求的上下文。
//
// 后续的中间件和处理函数通过 Context.Context() 获取到的将是新的上下文，
// 常用于超时控制、链路追踪等场景。传入 nil 时不做任何修改。
func (c *Context) UseContext(ctx context.Context) {
	if ctx != nil {
		c.Request = c.Request.WithContext(ctx)
	}
}

func (c *Context) Context() context.Context {
	return c.Request.Context()
}

func (c *Context) reset(writer http.ResponseWriter) {
	c.writer.reset(writer)
	c.Writer = &c.writer
	if c.Params == nil {
		c.Params = make(Params, 0, 8)
	}
	c.Params = c.Params[:0]
	c.index = -1
	c.aborted = false
	c.handlers = nil
	c.route = nil
	c.errorEncoder = nil
	clear(c.Keys)
}

func (c *Context) Next() {
	if c.route == nil || c.aborted {
		return
	}

	c.index++

	var globalHandlers = c.handlers
	for c.index < len(globalHandlers) && !c.aborted {
		globalHandlers[c.index](c)
		c.index++
	}

	var routeHandlers = c.route.Handlers()
	for c.index-len(globalHandlers) < len(routeHandlers) && !c.aborted {
		routeHandlers[c.index-len(globalHandlers)](c)
		c.index++
	}
}

func (c *Context) Route() Route {
	return c.route
}

func (c *Context) Param(name string) string {
	return c.Params.ByName(name)
}

func (c *Context) Aborted() bool {
	return c.aborted
}

func (c *Context) Abort() {
	c.aborted = true
}

func (c *Context) AbortWithError(err error) {
	c.Abort()
	c.Status(http.StatusInternalServerError)
	c.errorEncoder(c, err)
	c.Writer.WriteHeaderNow()
}

func (c *Context) AbortWithStatus(code int) {
	c.Abort()
	c.Status(code)
	c.Writer.WriteHeaderNow()
}

func (c *Context) AbortWithJSON(code int, obj any) {
	c.Abort()
	c.JSON(code, obj)
}

func (c *Context) JSON(code int, obj any) {
	c.Render(code, jsonRender{data: obj})
}

func (c *Context) String(code int, s string) {
	c.Render(code, textRender{text: s})
}

func (c *Context) Render(code int, r Render) {
	if r == nil {
		return
	}

	c.Status(code)

	var header = c.Writer.Header()
	if val := header[kContentType]; len(val) == 0 {
		header[kContentType] = r.ContentType()
	}

	if !bodyAllowedForStatus(code) {
		c.Writer.WriteHeaderNow()
		return
	}

	_ = r.Render(c.Writer)
	//if err := r.Render(c.Writer); err != nil {
	//	c.AbortWithError(err)
	//}
}

func (c *Context) Status(code int) {
	c.Writer.WriteHeader(code)
}

func (c *Context) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}
	c.Keys[key] = value
}

func (c *Context) Get(key string) (value any, exists bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists = c.Keys[key]
	return
}

func bodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == http.StatusNoContent:
		return false
	case status == http.StatusNotModified:
		return false
	}
	return true
}
