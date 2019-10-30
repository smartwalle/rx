package rx

import "net/http"

const (
	defaultWriteSize   = -1
	defaultWriteStatus = http.StatusOK
)

type ResponseWriter interface {
	http.ResponseWriter

	Writer() http.ResponseWriter

	StatusCode() int

	WriteHeaderNow()

	Written() bool
}

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (this *responseWriter) reset(w http.ResponseWriter) {
	this.ResponseWriter = w
	this.size = defaultWriteSize
	this.status = defaultWriteStatus
}

func (this *responseWriter) Writer() http.ResponseWriter {
	return this.ResponseWriter
}

func (this *responseWriter) StatusCode() int {
	return this.status
}

func (this *responseWriter) WriteHeader(statusCode int) {
	if statusCode > 0 && !this.Written() {
		this.status = statusCode
	}
}

func (this *responseWriter) WriteHeaderNow() {
	if !this.Written() {
		this.ResponseWriter.WriteHeader(this.status)
	}
}

func (this *responseWriter) Write(b []byte) (n int, err error) {
	this.WriteHeaderNow()
	n, err = this.ResponseWriter.Write(b)
	this.size = this.size + n
	return n, err
}

func (this *responseWriter) Written() bool {
	return this.size != defaultWriteSize
}
