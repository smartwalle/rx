package main

import (
	"fmt"
	"github.com/smartwalle/rx"
	"net/http"
)

func main() {
	var s = rx.New()
	s.Use(func(c *rx.Context) {
		fmt.Fprintln(c.Writer, "global m1")
	})
	s.Use(func(c *rx.Context) {
		fmt.Fprintln(c.Writer, "global m2")
	})
	s.GET("/hello", func(c *rx.Context) {
		fmt.Fprintln(c.Writer, "hello")
	})

	s.Use(func(c *rx.Context) {
		fmt.Fprintln(c.Writer, "global m3")
	})
	s.GET("/world", func(c *rx.Context) {
		fmt.Fprintln(c.Writer, "world")
	})

	http.ListenAndServe(":8893", s)
}
