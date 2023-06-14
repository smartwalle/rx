package main

import (
	"github.com/gin-gonic/gin"
	"github.com/smartwalle/rx"
	"log"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	var s = rx.New()
	var provider = rx.NewListProvider()
	provider.Add("/user", []string{"http://127.0.0.1:9910", "http://127.0.0.1:9911"})
	provider.Add("/order", []string{"http://127.0.0.1:9920", "http://127.0.0.1:9921"})
	provider.Add("/ws", []string{"http://127.0.0.1:9930", "http://127.0.0.1:9931"})
	s.Load(provider)

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
