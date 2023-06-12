package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	for i := 0; i < 1000000; i++ {
		var rsp, err = http.Get("http://127.0.0.1:9901/test")
		if err != nil {
			fmt.Println(err)
			return
		}
		rsp.Body.Close()
		time.Sleep(time.Millisecond * 10)
	}
}
