package main

import (
	"github.com/smartwalle/rx"
	"net/http"
)

func main() {
	var s = rx.New()

	s.GET("/test/*dd", []string{"http://127.0.0.1:9910", "http://127.0.0.1:9911"}, func(c *rx.Context) {
		c.Next()
	})
	http.ListenAndServe(":9900", s)
}
