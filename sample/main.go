package main

import (
	"github.com/smartwalle/rx"
	"net/http"
)

func main() {
	var s = rx.New()
	s.GET("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(request.URL.Path))
	})
	s.GET("/t1", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(request.URL.Path))
	})
	s.GET("/t1/h1", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(request.URL.Path))
	})
	s.GET("/t1/h2", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(request.URL.Path))
	})
	s.GET("/t2/h1", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(request.URL.Path))
	})
	s.GET("/t2/h2", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(request.URL.Path))
	})

	s.Print()

	http.ListenAndServe(":9987", s)

}
