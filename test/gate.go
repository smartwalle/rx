package main

//import (
//	"github.com/gin-gonic/gin"
//	"github.com/smartwalle/rx"
//	"net/http/httputil"
//	"net/url"
//)
//
//func main() {
//	var s = gin.Default()
//	var u, _ = url.Parse("http://127.0.0.1:9900")
//	ss := httputil.NewSingleHostReverseProxy(u)
//	s.GET("/test/*dd", func(c *gin.Context) {
//		c.Request.URL.Path = rx.CleanPath(c.Request.URL.Path)
//		ss.ServeHTTP(c.Writer, c.Request)
//	})
//	s.Run(":9901")
//}

import (
	"github.com/smartwalle/rx"
	"net/http"
)

func main() {
	var s = rx.New()

	s.GET("/test/*dd", []string{"http://127.0.0.1:9910", "http://127.0.0.1:9911"}, func(c *rx.Context) {
		c.Next()
	})
	http.ListenAndServe(":9901", s)
}
