package main

import (
	"fmt"
	"github.com/smartwalle/rx"
	"net/http"
)

func main() {
	var s = rx.New()
	s.Use(func(c *rx.Context) {
		var h = c.Writer.Header()
		h.Add("Access-Control-Allow-Origin", "*")
		h.Add("Access-Control-Allow-Credentials", "true")
		h.Add("Access-Control-Allow-Methods", "GET,POST,DELETE,PUT,OPTIONS")
		h.Add("Access-Control-Allow-Headers", "Sec-Websocket-Key, Connection, Sec-Websocket-Version, Sec-Websocket-Extensions, Upgrade, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})
	s.POST("/test", func(c *rx.Context) {
		fmt.Fprintln(c.Writer, "hello")
	})
	http.ListenAndServe(":8896", s)
}
