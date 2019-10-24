package rx

import (
	"net/http"
)

type HandlerFunc func(c *Context)

type Context struct {
	Request  *http.Request
	Writer   http.ResponseWriter
	handlers []HandlerFunc
	params   Params
	index    int
	abort    bool
}

func (this *Context) reset() {
	this.Request = nil
	this.Writer = nil
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

func (this *Context) Params() Params {
	return this.params
}

func (this *Context) Param(key string) string {
	return this.params[key]
}
