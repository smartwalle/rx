package rx

import (
	"log"
	"time"
)

func Logger() HandlerFunc {
	return func(c *Context) {
		var begin = time.Now()

		c.Next()

		log.Printf("| %3d | %13v | %s - %s \n", c.Writer.Status(), time.Since(begin), c.Request.Method, c.Request.URL.Path)
	}
}
