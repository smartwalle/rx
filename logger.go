package rx

import (
	"log"
	"time"
)

func Logger() HandlerFunc {
	return func(c *Context) {
		var begin = time.Now()
		var method = c.Request.Method
		var path = c.Request.URL.Path

		c.Next()

		log.Printf("| %3d | %13v | %s - %s \n", c.Writer.Status(), time.Since(begin), method, path)
	}
}
