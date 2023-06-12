package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	var s = gin.Default()
	s.GET("/test/login", func(c *gin.Context) {
		c.Writer.WriteString("9913")
	})
	s.Run(":9913")
}
