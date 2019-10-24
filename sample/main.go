package main

import (
	"github.com/smartwalle/rx"
	"net/http"
)

func main() {
	var s = rx.New()
	s.GET("/", func(c *rx.Context) {
		c.Writer.Write([]byte(c.Request.URL.Path))
	})
	s.GET("/t1", func(c *rx.Context) {
		c.Writer.Write([]byte(c.Request.URL.Path))
	})
	s.GET("/t1/h1", func(c *rx.Context) {
		c.Writer.Write([]byte(c.Request.URL.Path))
	})
	s.GET("/t1/h2", func(c *rx.Context) {
		c.Writer.Write([]byte(c.Request.URL.Path))
	})
	s.GET("/t2/h1", func(c *rx.Context) {
		c.Writer.Write([]byte(c.Request.URL.Path))
	})
	s.GET("/t2/h2", func(c *rx.Context) {
		c.Writer.Write([]byte(c.Request.URL.Path))
	})

	s.Print()

	http.ListenAndServe(":9987", s)

}
