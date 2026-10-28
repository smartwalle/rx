package main

import (
	"net/http"

	"github.com/smartwalle/rx"
	"github.com/smartwalle/rx/router"
)

func main() {
	var r = router.New()
	r.GET("/r1", func(c *rx.Context) {
		c.JSON(http.StatusOK, "r1")
	})

	r.GET("/r2", func(c *rx.Context) {
		c.JSON(http.StatusOK, "r2")
	})

	r.GET("/r3", func(c *rx.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "https://www.baidu.com")
	})

	var s = rx.New()
	s.Use(rx.Logger())
	s.UseRouter(r)
	s.UseErrorEncoder(func(c *rx.Context, err error) {
		c.String(http.StatusBadRequest, err.Error())
	})
	s.UseNotFoundHandler(func(c *rx.Context) {
		c.String(http.StatusNotFound, "访问路由不存在")
	})

	http.ListenAndServe(":9900", s)
}
