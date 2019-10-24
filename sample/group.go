package main

import (
	"fmt"
	"github.com/smartwalle/rx"
	"net/http"
)

func main() {

	var s = rx.New()

	s.Use(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Println("第一个 middleware")
	})
	s.Use(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Println("第二个 middleware")
	})

	var user = s.Group("/user", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Println("Group user 的第一个 middleware")
	})
	user.Use(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Println("Group user 的第二个 middleware")
	})
	user.GET("/list", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Println("用户列表")
	})

	s.Use(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Println("第三个 middleware")
	})
	var order = s.Group("/order", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Println("Group order 的第一个 middleware")
	})
	order = order.Use(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Println("Group order 的第二个 middleware")
	})
	order.GET("/list", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Println("订单列表")
	})

	http.ListenAndServe(":9986", s)
}
