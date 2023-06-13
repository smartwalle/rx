package main

import (
	"github.com/smartwalle/rx"
	"log"
	"net/http"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	var s = rx.New()
	s.Add("/user", []string{"http://127.0.0.1:9910", "http://127.0.0.1:9911"})
	s.Add("/order", []string{"http://127.0.0.1:9920", "http://127.0.0.1:9921"})
	s.Add("/ws", []string{"http://127.0.0.1:9930", "http://127.0.0.1:9931"})

	http.Handle("/", s)
	// or
	// http.HandleFunc("/", s.ServeHTTP)
	http.ListenAndServe(":9901", nil)
}
