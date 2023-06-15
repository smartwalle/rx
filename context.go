package rx

import (
	"math"
	"net/http"
	"net/http/httputil"
)

const (
	kContentType      = "Content-Type"
	kAbortIndex  int8 = math.MaxInt8 >> 1
)

type Context struct {
	mWriter responseWriter
	Request *http.Request
	Writer  ResponseWriter

	index int8

	handlers HandlersChain
	proxy    *httputil.ReverseProxy
	Route    *Route
}

func (c *Context) reset() {
	c.Writer = &c.mWriter
	c.index = -1
	c.handlers = nil
	c.proxy = nil
	c.Route = nil
}

func (c *Context) Next() {
	if c.Route != nil {
		c.index++

		var hLen = int8(len(c.handlers))
		for c.index < hLen {
			c.handlers[c.index](c)
			c.index++
		}

		for c.index-hLen < int8(len(c.Route.handlers)) {
			c.Route.handlers[c.index-hLen](c)
			c.index++
		}

		hLen = int8(len(c.handlers) + len(c.Route.handlers))
		if c.index-hLen < 1 && c.proxy != nil {
			c.proxy.ServeHTTP(c.Writer, c.Request)
			c.index++
		}
	}
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
