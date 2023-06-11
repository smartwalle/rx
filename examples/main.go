package main

import (
	"github.com/smartwalle/rx"
	"net/http"
)

func main() {
	var s = rx.New()
	s.GET("/hello", func(c *rx.Context) {
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Write([]byte(c.Request.URL.Path))
	})
	s.GET("/world", func(c *rx.Context) {
		c.Write(http.StatusOK, []byte(c.Request.URL.Path))
	})
	s.NotFound(func(c *rx.Context) {
		c.Writer.Write([]byte("什么?"))
	})

	http.ListenAndServe(":8891", s)
}
