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

		var now = time.Now()
		var statusCode = c.Writer.Status()

		log.Printf("| %3d | %13v | %s - %s \n", statusCode, now.Sub(begin), method, path)
	}
}
