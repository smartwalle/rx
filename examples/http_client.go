package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

func main() {
	var begin = time.Now()
	var wait = &sync.WaitGroup{}

	for i := 0; i < 10000; i++ {
		wait.Add(3)
		go request(wait)
		go request(wait)
		go request(wait)
		time.Sleep(time.Millisecond * 10)
	}

	wait.Wait()
	fmt.Println(time.Now().Sub(begin))
}

func request(wait *sync.WaitGroup) {
	defer wait.Done()
	var rsp, err = http.Get("http://127.0.0.1:9900/user/list")
	if err != nil {
		log.Println(err)
		return
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		log.Println(rsp.StatusCode)
		return
	}

	var data, _ = io.ReadAll(rsp.Body)
	if len(data) == 0 {
		log.Println(string(data))
		return
	}
}
