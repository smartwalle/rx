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
	s.NoRoute(func(c *rx.Context) {
		fmt.Fprintln(c.Writer, "not found")
	})

	// user
	var user = s.Group("/user", func(c *rx.Context) {
		fmt.Fprintln(c.Writer, "user m1")
	})
	user.Use(func(c *rx.Context) {
		fmt.Fprintln(c.Writer, "user m2")
	})
	user.GET("/list", func(c *rx.Context) {
		fmt.Fprintln(c.Writer, "user list")
	})
	user.GET("/detail", func(c *rx.Context) {
		fmt.Fprintln(c.Writer, "user detail")
	})

	// order
	var order = s.Group("/order")
	order.Use(func(c *rx.Context) {
		fmt.Fprintln(c.Writer, "order m1")
	}, func(c *rx.Context) {
		fmt.Fprintln(c.Writer, "order m2")
	})
	order.GET("/list", func(c *rx.Context) {
		fmt.Fprintln(c.Writer, "order list")
	})
	order.GET("/:id", func(c *rx.Context) {
		fmt.Fprintln(c.Writer, "order detail")
		fmt.Fprintln(c.Writer, "order id", c.Param("id"))
	})

	// point
	s.Group("/point", func(c *rx.Context) {
		fmt.Fprintln(c.Writer, "nothing to do")
	})

	http.ListenAndServe(":8892", s)
}
