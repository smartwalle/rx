package main

import (
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	var begin = time.Now()
	var wait = &sync.WaitGroup{}

	for i := 0; i < 10000; i++ {
		wait.Add(6)

		go request(wait, "http://127.0.0.1:9900/user/list")
		go request(wait, "http://127.0.0.1:9900/user/list")
		go request(wait, "http://127.0.0.1:9900/user/list")
		go request(wait, "http://127.0.0.1:9900/order/list")
		go request(wait, "http://127.0.0.1:9900/order/list")
		go request(wait, "http://127.0.0.1:9900/order/list")
		time.Sleep(time.Millisecond * 10)
	}

	wait.Wait()
	log.Println(time.Now().Sub(begin))
}

func request(wait *sync.WaitGroup, target string) {
	defer wait.Done()

	var rsp, err = http.Get(target)
	if err != nil {
		log.Println(err)
		return
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		log.Println(rsp.StatusCode)
		return
	}

	data, err := io.ReadAll(rsp.Body)
	if len(data) == 0 {
		log.Println(rsp.StatusCode, err)
		return
	}
}
