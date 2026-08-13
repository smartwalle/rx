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

	var s = rx.New()
	s.Use(rx.Logger())
	s.UseRouter(r)
	s.NoRoute(func(c *rx.Context) {
		c.JSON(http.StatusOK, "sorry")
	})

	http.ListenAndServe(":9900", s)
}
