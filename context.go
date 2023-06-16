package rx

import (
	"fmt"
	"math"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

const (
	kContentType      = "Content-Type"
	kAbortIndex  int8 = math.MaxInt8 >> 1
)

type rWriterWrapper struct {
	ResponseWriter
	context *Context
}

type Context struct {
	mWriter responseWriter
	Request *http.Request
	Writer  ResponseWriter

	index int8

	handlers HandlersChain
	proxy    *httputil.ReverseProxy
	target   *url.URL
	route    *Route

	mu   sync.RWMutex
	Keys map[string]interface{}
}

func (c *Context) reset() {
	c.Writer = &c.mWriter
	c.index = -1
	c.handlers = nil
	c.proxy = nil
	c.target = nil
	c.route = nil
	c.Keys = nil
}

func (c *Context) Next() {
	if c.route != nil {
		c.index++

		var hLen = int8(len(c.handlers))
		for c.index < hLen {
			c.handlers[c.index](c)
			c.index++
		}

		for c.index-hLen < int8(len(c.route.handlers)) {
			c.route.handlers[c.index-hLen](c)
			c.index++
		}

		hLen = int8(len(c.handlers) + len(c.route.handlers))
		if c.index-hLen < 1 && c.proxy != nil {
			c.proxy.ServeHTTP(rWriterWrapper{ResponseWriter: c.Writer, context: c}, c.Request)
			c.index++
		}
	}
}

func (c *Context) Error(err error) {
	// TODO abort
	fmt.Println(err)
}

func (c *Context) Target() *url.URL {
	return c.target
}

func (c *Context) Route() *Route {
	return c.route
}

func (c *Context) IsAborted() bool {
	return c.index >= kAbortIndex
}

func (c *Context) Abort() {
	c.index = kAbortIndex
}

func (c *Context) AbortWithStatus(code int) {
	c.Status(code)
	c.Writer.WriteHeaderNow()
	c.Abort()
}

func (c *Context) AbortWithJSON(code int, obj interface{}) {
	c.Abort()
	c.JSON(code, obj)
}

func (c *Context) JSON(code int, obj interface{}) {
	c.Render(code, jsonRender{data: obj})
}

func (c *Context) String(code int, s string) {
	c.Render(code, textRender{text: s})
}

func (c *Context) Render(code int, r Render) {
	if r == nil {
		return
	}

	c.Writer.WriteHeader(code)

	var header = c.Writer.Header()
	if val := header[kContentType]; len(val) == 0 {
		header[kContentType] = r.ContentType()
	}

	if !bodyAllowedForStatus(code) {
		c.Writer.WriteHeaderNow()
		return
	}

	if err := r.Render(c.Writer); err != nil {
		c.Abort()
	}
}

func (c *Context) Status(code int) {
	c.Writer.WriteHeader(code)
}

func (c *Context) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}
	c.Keys[key] = value
}

func (c *Context) Get(key string) (value interface{}, exists bool) {
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
