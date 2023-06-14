package rx

import (
	"math"
	"net/http"
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
	Location *Location
}

func (c *Context) reset(writer http.ResponseWriter, request *http.Request, handlers HandlersChain) {
	c.mWriter.reset(writer)
	c.Writer = &c.mWriter
	c.Request = request
	c.index = -1
	c.handlers = handlers
	c.Location = nil
}

func (c *Context) Next() {
	if c.Location != nil {
		c.index++

		var hLen = int8(len(c.handlers))
		for c.index < hLen {
			c.handlers[c.index](c)
			c.index++
		}

		for c.index-hLen < int8(len(c.Location.handlers)) {
			c.Location.handlers[c.index-hLen](c)
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
	c.Render(code, JSONRender{data: obj})
}

func (c *Context) String(code int, s string) {
	c.Render(code, TextRender{text: s})
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
