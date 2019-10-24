package rx

import (
	"net/http"
)

type HandlerFunc func(c *Context)

type Context struct {
	Request  *http.Request
	Writer   http.ResponseWriter
	handlers []HandlerFunc
	index    int
	abort    bool
}

func (this *Context) reset() {
	this.index = -1
	this.abort = false
	this.handlers = nil
	this.Request = nil
	this.Writer = nil
}

func (this *Context) Next() {
	this.index++
	for !this.abort && this.index < len(this.handlers) {
		var h = this.handlers[this.index]
		h(this)
		this.index++
	}
}

func (this *Context) Abort() {
	this.abort = true
}
