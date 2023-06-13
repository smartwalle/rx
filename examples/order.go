package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	RunOrderServer("9920", "9921")
	select {}
}

func RunOrderServer(ports ...string) {
	for _, port := range ports {
		go orderServer(port)
	}
}

func orderServer(port string) {
	var s = gin.Default()

	s.GET("/order/:id", func(context *gin.Context) {
		context.Writer.WriteString(fmt.Sprintf("从【%s】获取【订单 %s】的信息", port, context.Param("id")))
	})

	s.GET("/order/list", func(context *gin.Context) {
		context.Writer.WriteString(fmt.Sprintf("从【%s】获取【订单列表】", port))
	})

	s.Run(":" + port)
}
