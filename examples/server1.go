package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
)

func main() {
	var s = gin.Default()
	s.GET("/test", func(c *gin.Context) {
		c.Writer.WriteString("9910")
	})
	s.GET("/test/hello", func(c *gin.Context) {
		c.Writer.WriteString("hello from 9910")
	})

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	s.GET("/ws", func(context *gin.Context) {
		var c, err = upgrader.Upgrade(context.Writer, context.Request, nil)
		if err != nil {
			return
		}

		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				return
			}
			fmt.Println(string(msg))
		}
	})
	s.Run(":9910")
}
