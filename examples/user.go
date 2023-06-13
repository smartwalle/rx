package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func main() {
	//var s = gin.Default()
	//s.GET("/test", func(c *gin.Context) {
	//	c.Writer.WriteString("9910")
	//})
	//s.GET("/test/hello", func(c *gin.Context) {
	//	c.Writer.WriteString("hello from 9910")
	//})
	//
	//var upgrader = websocket.Upgrader{
	//	ReadBufferSize:  1024,
	//	WriteBufferSize: 1024,
	//}
	//upgrader.CheckOrigin = func(r *http.Request) bool {
	//	return true
	//}
	//s.GET("/ws", func(context *gin.Context) {
	//	var c, err = upgrader.Upgrade(context.Writer, context.Request, nil)
	//	if err != nil {
	//		return
	//	}
	//
	//	for {
	//		_, msg, err := c.ReadMessage()
	//		if err != nil {
	//			return
	//		}
	//		fmt.Println(string(msg))
	//	}
	//})
	//s.Run(":9910")

	RunUserServer("9910", "9911")
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
		context.Writer.WriteString(fmt.Sprintf("从【%s】获取【用户列表】", port))
	})

	s.Run(":" + port)
}
