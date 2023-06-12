package main

import (
	"github.com/gin-gonic/gin"
	"github.com/smartwalle/rx"
)

func main() {
	var s = rx.New()
	s.Add("/test/login", []string{"http://127.0.0.1:9913"})
	s.Add("/test", []string{"http://127.0.0.1:9910", "http://127.0.0.1:9911"})

	var server = gin.Default()
	server.NoRoute(func(context *gin.Context) {
		s.ServeHTTP(context.Writer, context.Request)
	})

	server.GET("/hi", func(context *gin.Context) {
		context.Writer.WriteString("hi from gate")
	})
	server.Run(":9901")
}
