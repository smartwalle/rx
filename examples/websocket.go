package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	RunWebsocketServer("9930", "9931")
	select {}
}

func RunWebsocketServer(ports ...string) {
	for _, port := range ports {
		go websocketServer(port)
	}
}

func websocketServer(port string) {
	var s = gin.Default()

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	s.GET("/ws", func(context *gin.Context) {
		var conn, err = upgrader.Upgrade(context.Writer, context.Request, nil)
		if err != nil {
			log.Println(port, "建立 Websocket 发生错误：", err)
			return
		}

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println(port, "接收消息发生错误：", err)
				return
			}
			log.Println(port, "收到消息：", string(msg))
		}
	})

	s.Run(":" + port)
}
