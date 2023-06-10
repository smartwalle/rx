package main

import (
	"github.com/smartwalle/rx"
	"net/http"
)

func main() {
	var s = rx.New()
	s.GET("/json", func(c *rx.Context) {
		c.JSON(http.StatusOK, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0})
	})
	http.ListenAndServe(":8897", s)
}
