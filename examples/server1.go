package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	var s = gin.Default()
	s.GET("/test", func(c *gin.Context) {
		c.Writer.WriteString("9910")
	})
	s.GET("/test/h", func(c *gin.Context) {
		c.Writer.WriteString("h")
	})
	s.Run(":9910")
}
