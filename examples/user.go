package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
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

	s.Run(":" + port)
}
