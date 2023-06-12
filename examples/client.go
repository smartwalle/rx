package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"time"
)

func main() {
	//for i := 0; i < 1000000; i++ {
	//	var rsp, err = http.Get("http://127.0.0.1:9901/test")
	//	if err != nil {
	//		fmt.Println(err)
	//		return
	//	}
	//	rsp.Body.Close()
	//	time.Sleep(time.Millisecond * 10)
	//}

	var c, _, err = websocket.DefaultDialer.Dial("ws://127.0.0.1:9910/ws", nil)
	fmt.Println(err)
	for {
		c.WriteMessage(websocket.TextMessage, []byte("sss"))
		time.Sleep(time.Second)
	}
}
