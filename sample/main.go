package main

import (
	"github.com/smartwalle/rx"
	"net/http"
)

func main() {
	var r = rx.NewRouter()
	r.GET("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(request.URL.Path))
	})
	r.GET("/t1", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(request.URL.Path))
	})
	r.GET("/t1/h1", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(request.URL.Path))
	})
	r.GET("/t1/h2", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(request.URL.Path))
	})
	r.GET("/t2/h1", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(request.URL.Path))
	})
	r.GET("/t2/h2", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(request.URL.Path))
	})
	r.Print()

	http.ListenAndServe(":9987", r)

}
