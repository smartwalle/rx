package main

import (
	"fmt"
	"github.com/smartwalle/rx"
	"net/http"
)

func main() {
	var s = rx.New()
	s.POST("/router/add", func(c *rx.Context) {
		c.Request.ParseForm()
		var path = c.Request.Form.Get("path")
		var text = c.Request.Form.Get("text")
		newRouter(s, path, text)
		c.Writer.Write([]byte(fmt.Sprintf("路由 %s 添加成功", path)))
	})
	s.POST("/router/del", func(c *rx.Context) {
		c.Request.ParseForm()
		var path = c.Request.Form.Get("path")
		delRouter(s, path)
	})
	http.ListenAndServe(":8895", s)

}

func newRouter(r *rx.Engine, path, text string) {
	var h = rx.HandlerFunc(func(c *rx.Context) {
		c.Writer.Write([]byte(text))
	})
	r.GET(path, h)
}

func delRouter(r *rx.Engine, path string) {
	r.Break(http.MethodGet, path)
}
