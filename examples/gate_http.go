package main

import (
	"github.com/smartwalle/rx"
	"log"
	"net/http"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	var provider = rx.NewListProvider()
	provider.Add("/user", []string{"http://127.0.0.1:9910", "http://127.0.0.1:9911"})
	provider.Add("/order", []string{"http://127.0.0.1:9920", "http://127.0.0.1:9921"})
	provider.Add("/ws", []string{"http://127.0.0.1:9930", "http://127.0.0.1:9931"})

	var s = rx.New()
	s.Load(provider)

	s.Use(func(c *rx.Context) {
		log.Println("middleware 1")
	})

	s.Use(func(c *rx.Context) {
		log.Println("middleware 2")
	})

	s.NoRoute(func(c *rx.Context) {
		log.Println("no route:", c.Request.URL.Path)
	})

	http.ListenAndServe(":9901", s)
}
