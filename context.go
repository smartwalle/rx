package rx

import (
	"net/http"
)

const (
	contentType = "Content-Type"
)

type Context struct {
	Request  *http.Request
	Writer   http.ResponseWriter
	handlers HandlerChain
	params   Params
	index    int
	abort    bool
}

func (this *Context) reset(w http.ResponseWriter, req *http.Request) {
	this.Request = req
	if this.Writer == nil {
		this.Writer = &responseWriter{}
	}
	this.Writer.(*responseWriter).reset(w)
	this.handlers = nil
	this.params = nil
	this.index = -1
	this.abort = false
}

func (this *Context) Next() {
	this.index++
	for !this.abort && this.index < len(this.handlers) {
		var handler = this.handlers[this.index]
		handler(this)
		this.index++
	}
}

func (this *Context) Abort() {
	this.abort = true
}

func (this *Context) AbortWithStatus(statusCode int) {
	this.abort = true
	this.Writer.WriteHeader(statusCode)
}

func (this *Context) Write(statusCode int, b []byte) {
	this.Writer.WriteHeader(statusCode)
	this.Writer.Write(b)
}

func (this *Context) Render(statusCode int, r Render) {
	if r == nil {
		return
	}

	this.Writer.WriteHeader(statusCode)

	var header = this.Writer.Header()
	if val := header[contentType]; len(val) == 0 {
		header[contentType] = r.ContentType()
	}

	if !bodyAllowedForStatus(statusCode) {
		var w = this.Writer.(*responseWriter)
		w.WriteHeaderNow()
	}

	if err := r.Render(this.Writer); err != nil {
		panic(err)
	}
}

func (this *Context) Params() Params {
	return this.params
}

func (this *Context) Param(key string) string {
	if this.params == nil {
		return ""
	}
	return this.params[key]
}
