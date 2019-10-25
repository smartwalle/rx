package rx

import "net/http"

const (
	defaultWriteSize   = -1
	defaultWriteStatus = http.StatusOK
)

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

func (this *responseWriter) StatusCode() int {
	return this.status
}

func (this *responseWriter) WriteHeader(statusCode int) {
	if statusCode > 0 && !this.Written() {
		this.status = statusCode
	}
}

func (this *responseWriter) WriteStatus() {
	if !this.Written() {
		this.ResponseWriter.WriteHeader(this.status)
	}
}

func (this *responseWriter) Write(b []byte) (n int, err error) {
	this.WriteStatus()
	n, err = this.ResponseWriter.Write(b)
	this.size = n
	return n, err
}

func (this *responseWriter) Written() bool {
	return this.size != defaultWriteSize
}
