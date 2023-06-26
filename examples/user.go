package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	RunUserServer("9910", "9911", "9912", "9913")
	select {}
}

func RunUserServer(ports ...string) {
	for _, port := range ports {
		go userServer(port)
	}
}

func userServer(port string) {
	var s = gin.Default()

	s.GET("/user/:id", func(context *gin.Context) {
		context.Writer.WriteString(fmt.Sprintf("从【%s】获取【用户 %s】的信息", port, context.Param("id")))
	})

	s.GET("/user/list", func(context *gin.Context) {
		log.Println(port)
		context.Writer.WriteString(fmt.Sprintf("从【%s】获取【用户列表】", port))
	})

	s.GET("/user/sse", func(context *gin.Context) {
		context.Writer.Header().Add("Content-Type", "text/event-stream")
		context.Writer.Header().Add("Cache-Control", "no-cache")
		context.Writer.Header().Add("Connection", "keep-alive")

		var idx = 0
		for {
			select {
			case <-context.Request.Context().Done():
				return
			default:
				context.Writer.WriteString(fmt.Sprintf("sse %d\n", idx))
				context.Writer.Flush()
				time.Sleep(time.Second)
				idx++
			}
		}
	})

	s.GET("/user/chunk", func(context *gin.Context) {
		context.Writer.Header().Add("Content-Type", "text/plain")
		context.Writer.Header().Add("Transfer-Encoding", "chunked")
		context.Writer.Header().Add("X-Content-Type-Options", "nosniff")

		var idx = 0
		for {
			select {
			case <-context.Request.Context().Done():
				return
			default:
				context.Writer.WriteString(fmt.Sprintf("chunk %d\n", idx))
				context.Writer.Flush()
				time.Sleep(time.Second)
				idx++
			}
		}
	})

	s.Run(":" + port)
}
