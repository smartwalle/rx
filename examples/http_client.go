package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	for i := 0; i < 10000; i++ {
		var rsp, err = http.Get("http://127.0.0.1:9900/user/list")
		if err != nil {
			log.Println(err)
			return
		}
		rsp.Body.Close()
		time.Sleep(time.Millisecond * 10)
	}
}
