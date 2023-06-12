package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	var s = gin.Default()
	s.GET("/test", func(c *gin.Context) {
		c.Writer.WriteString("9911")
	})
	s.GET("/test/hello", func(c *gin.Context) {
		c.Writer.WriteString("hello from 9911")
	})
	s.Run(":9911")
}
