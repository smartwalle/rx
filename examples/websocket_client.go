package main

import (
	"github.com/gorilla/websocket"
	"log"
	"time"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	var conn, _, err = websocket.DefaultDialer.Dial("ws://127.0.0.1:9900/ws", nil)
	if err != nil {
		log.Println(err)
		return
	}
	for {
		conn.WriteMessage(websocket.TextMessage, []byte("sss"))
		time.Sleep(time.Second)
	}
}
