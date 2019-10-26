package main

import (
	"fmt"
	"github.com/smartwalle/rx"
	"net/http"
)

func main() {
	var s = rx.New()
	s.GET("/normal", func(c *rx.Context) {
		fmt.Fprintln(c.Writer, "normal")
	})
	s.GET("/user/list", func(c *rx.Context) {
		fmt.Fprintln(c.Writer, "user list")
	})
	s.GET("/user/:id", func(c *rx.Context) {
		fmt.Fprintln(c.Writer, "user detail 1")
		fmt.Fprintln(c.Writer, "user id", c.Param("id"))
	})
	s.GET("/user/:id/{age:([\\d]+)}", func(c *rx.Context) {
		fmt.Fprintln(c.Writer, "user detail 2")
		fmt.Fprintln(c.Writer, "user id", c.Param("id"))
		fmt.Fprintln(c.Writer, "user age", c.Param("age"))
	})

	s.GET("/user/:id/order", func(c *rx.Context) {
		fmt.Fprintln(c.Writer, "order list")
		fmt.Fprintln(c.Writer, "user id", c.Param("id"))
	})
	s.GET("/user/:id/order/:order_id", func(c *rx.Context) {
		fmt.Fprintln(c.Writer, "order detail")
		fmt.Fprintln(c.Writer, "user id", c.Param("id"))
		fmt.Fprintln(c.Writer, "order id", c.Param("order_id"))
	})

	s.GET("/user/*path", func(c *rx.Context) {
		fmt.Fprintln(c.Writer, "/user/*path")
		fmt.Fprintln(c.Writer, "path", c.Param("path"))
	})

	http.ListenAndServe(":8894", s)
}
