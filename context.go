package rx

import (
	"fmt"
	"github.com/smartwalle/rx/balancer"
	"net/http"
)

const (
	kContentType = "Content-Type"
)

type Context struct {
	Request       *http.Request
	Writer        ResponseWriter
	defaultWriter *responseWriter
	handlers      HandlersChain
	balancer      balancer.Balancer
	params        Params
	index         int
	abort         bool
}

func newContext() *Context {
	return &Context{defaultWriter: &responseWriter{}}
}

func (this *Context) reset(w http.ResponseWriter, req *http.Request) {
	this.Request = req
	this.defaultWriter.reset(w)
	this.Writer = this.defaultWriter
	this.handlers = nil
	this.balancer = nil
	this.params = this.params[0:0]
	this.index = -1
	this.abort = false
}

func (this *Context) Next() {
	this.index++
	for !this.abort && this.index < len(this.handlers) {
		this.handlers[this.index](this)
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

func (this *Context) JSON(statusCode int, obj interface{}) {
	this.Render(statusCode, JSONRender{data: obj})
}

func (this *Context) Render(statusCode int, r Render) {
	if r == nil {
		return
	}

	this.Writer.WriteHeader(statusCode)

	var header = this.Writer.Header()
	if val := header[kContentType]; len(val) == 0 {
		header[kContentType] = r.ContentType()
	}

	if !bodyAllowedForStatus(statusCode) {
		this.Writer.WriteHeaderNow()
		return
	}

	if err := r.Render(this.Writer); err != nil {
		panic(err)
	}
}

func (this *Context) Params() Params {
	return this.params
}

func (this *Context) Param(key string) string {
	return this.params.ByName(key)
}

func (this *Context) proxy() {
	if !this.abort {
		var target, err = this.balancer.Pick(this.Request)
		if err != nil {
			this.error(http.StatusBadGateway, []byte(fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadGateway), err.Error())))
			return
		}
		this.Request.URL.Path = CleanPath(this.Request.URL.Path)
		target.ServeHTTP(this.Writer, this.Request)
	}
}

func (this *Context) exec() {
	this.Next()
	this.proxy()
	this.Writer.WriteHeaderNow()
}

func (this *Context) error(statusCode int, body []byte) {
	this.Writer.WriteHeader(statusCode)
	if this.Writer.Written() {
		return
	}
	this.Writer.Write(body)
}
