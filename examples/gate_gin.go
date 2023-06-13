package main

import (
	"github.com/gin-gonic/gin"
	"github.com/smartwalle/rx"
	"log"
	"time"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	var s = rx.New()
	s.Add("/user", []string{"http://127.0.0.1:9910", "http://127.0.0.1:9911"})
	s.Add("/order", []string{"http://127.0.0.1:9920", "http://127.0.0.1:9921"})
	s.Add("/ws", []string{"http://127.0.0.1:9930", "http://127.0.0.1:9931"})

	go func() {
		var idx = 0
		var uList = [][]string{{"http://127.0.0.1:9910"}, {"http://127.0.0.1:9911"}}
		for {
			time.Sleep(time.Second * 3)
			idx += 1

			time.Sleep(time.Second)
			var location, _ = s.BuildLocation("/user", uList[idx%2])
			s.UpdateLocations([]*rx.Location{location})
		}
	}()

	go func() {
		var idx = 0
		var uList = [][]string{{"http://127.0.0.1:9912"}, {"http://127.0.0.1:9913"}}
		for {
			time.Sleep(time.Second * 3)
			idx += 1

			time.Sleep(time.Second)
			var location, _ = s.BuildLocation("/user", uList[idx%2])
			s.UpdateLocations([]*rx.Location{location})
		}
	}()

	var gate = gin.New()
	gate.Any("/user/*xx", func(context *gin.Context) {
		s.ServeHTTP(context.Writer, context.Request)
	})
	gate.Any("/order/*xx", func(context *gin.Context) {
		s.ServeHTTP(context.Writer, context.Request)
	})

	gate.Any("/ws", func(context *gin.Context) {
		s.ServeHTTP(context.Writer, context.Request)
	})

	gate.GET("/gate", func(context *gin.Context) {
		context.Writer.WriteString("来自网关的消息")
	})
	gate.Run(":9900")
}
