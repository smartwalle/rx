package main

import (
	"log"
	"net/http"

	"github.com/smartwalle/rx"
	"github.com/smartwalle/rx/router"
	"github.com/smartwalle/rx/staticfile"
)

func main() {
	var r = router.New()
	r.GET("/r1", func(c *rx.Context) {
		c.JSON(http.StatusOK, "r1")
	})

	r.GET("/r2", func(c *rx.Context) {
		c.JSON(http.StatusOK, "r2")
	})

	r.GET("/baidu", func(c *rx.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "https://www.baidu.com")
	})

	r.GET("/static/*filepath", staticfile.NewHandler("./"))

	var s = rx.New()
	s.Use(rx.Logger())
	s.UseRouter(r)
	s.UseErrorEncoder(func(c *rx.Context, err error) {
		c.String(http.StatusBadRequest, err.Error())
	})
	s.UseNotFoundHandler(func(c *rx.Context) {
		c.String(http.StatusNotFound, "访问路由不存在")
	})

	if err := http.ListenAndServe(":9900", s); err != nil {
		log.Fatalln(err)
	}
}
