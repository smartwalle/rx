package main

import (
	"github.com/smartwalle/rx"
	"net/http"
)

func main() {
	var s = rx.New()
	s.Add("/test/login", []string{"http://127.0.0.1:9913"})
	s.Add("/test", []string{"http://127.0.0.1:9910", "http://127.0.0.1:9911"})

	http.Handle("/", s)
	// or
	// http.HandleFunc("/", s.ServeHTTP)
	http.ListenAndServe(":9902", nil)
}
