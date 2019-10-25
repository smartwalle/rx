package main

import (
	"fmt"
	"github.com/smartwalle/rx"
	"net/http"
)

func main() {

	var s = rx.New()

	s.Use(func(c *rx.Context) {
		fmt.Println("第一个 middleware")
	})
	s.Use(func(c *rx.Context) {
		fmt.Println("第二个 middleware")
	})

	var user = s.Group("/user", func(c *rx.Context) {
		fmt.Println("Group user 的第一个 middleware")
	})
	user.Use(func(c *rx.Context) {
		fmt.Println("Group user 的第二个 middleware")
	})
	user.GET("/list", func(c *rx.Context) {
		fmt.Println("用户列表")
		c.Writer.Write([]byte(c.Request.URL.Path))
	})
	user.GET("/detail", func(c *rx.Context) {
		fmt.Println("用户详情")
		c.Writer.Write([]byte(c.Request.URL.Path))
	})

	s.Use(func(c *rx.Context) {
		fmt.Println("第三个 middleware")
	})
	var order = s.Group("/order", func(c *rx.Context) {
		fmt.Println("Group order 的第一个 middleware")
	})
	order.Use(func(c *rx.Context) {
		fmt.Println("Group order 的第二个 middleware")
	})
	order.GET("/list", func(c *rx.Context) {
		fmt.Println("订单列表")
		c.Writer.Write([]byte(c.Request.URL.Path))
	})
	order.GET("/detail", func(c *rx.Context) {
		fmt.Println("Group order 的第三个 middleware, 不会继续执行后续 handler")
		c.Abort()
	}, func(c *rx.Context) {
		fmt.Println("订单详情")
		c.Writer.Write([]byte(c.Request.URL.Path))
	})
	order.GET("/detail/:id", func(c *rx.Context) {
		fmt.Println("订单详情x")
		c.Writer.Write([]byte(c.Request.URL.Path))
	})

	http.ListenAndServe(":9986", s)
}
