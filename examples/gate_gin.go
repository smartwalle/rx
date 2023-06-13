package main

import (
	"github.com/gin-gonic/gin"
	"github.com/smartwalle/rx"
)

func main() {
	var s = rx.New()
	s.Add("/user", []string{"http://127.0.0.1:9910", "http://127.0.0.1:9911"})
	s.Add("/order", []string{"http://127.0.0.1:9920", "http://127.0.0.1:9921"})
	s.Add("/ws", []string{"http://127.0.0.1:9930", "http://127.0.0.1:9931"})

	var gate = gin.Default()
	gate.Any("/user/*xx", func(context *gin.Context) {
		s.ServeHTTP(context.Writer, context.Request)
	})
	gate.Any("/order/*xx", func(context *gin.Context) {
		s.ServeHTTP(context.Writer, context.Request)
	})

	gate.Any("/ws", func(context *gin.Context) {
		s.ServeHTTP(context.Writer, context.Request)
	})

	gate.GET("/gate", func(context *gin.Context) {
		context.Writer.WriteString("来自网关的消息")
	})
	gate.Run(":9900")
}
