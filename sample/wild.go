package main

import (
	"fmt"
	"github.com/smartwalle/rx"
	"net/http"
)

func main() {
	var s = rx.New()
	s.Use(rx.Log())
	s.GET("/normal", func(c *rx.Context) {
		fmt.Fprintln(c.Writer, "normal")
	})

	s.GET("/user/:id", func(c *rx.Context) {
		fmt.Fprintln(c.Writer, "user detail 1")
		fmt.Fprintln(c.Writer, "user id", c.Param("id"))
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

	http.ListenAndServe(":8894", s)
}
