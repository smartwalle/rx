package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/smartwalle/rx"
	"log"
	"net/http"
	"regexp"
)

type gateGinRouter struct {
	routes []gateGinRoute
}

type gateGinRoute struct {
	regexp *regexp.Regexp
	route  rx.Route
}

func (r *gateGinRouter) Add(pattern string, targets []string, opts ...rx.Option) error {
	route, err := rx.NewProxyRoute(pattern, targets, opts...)
	if err != nil {
		return err
	}
	nRegexp, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	r.routes = append(r.routes, gateGinRoute{regexp: nRegexp, route: route})
	return nil
}

func (r *gateGinRouter) Match(req *http.Request) (rx.Route, error) {
	for _, route := range r.routes {
		if route.regexp.MatchString(req.URL.Path) {
			return route.route, nil
		}
	}
	return nil, nil
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	var rHandler = func(c *rx.Context) {
		defer func() {
			if v := recover(); v != nil {
				if err, ok := v.(error); ok {
					if errors.Is(err, http.ErrAbortHandler) {
						return
					}
				}
				panic(v)
			}
		}()
		c.Next()
	}

	var provider = &gateGinRouter{}
	provider.Add("/user/sse", []string{"http://127.0.0.1:9910", "http://127.0.0.1:9911"}, rx.WithHandlers(rHandler))
	provider.Add("/user/chunk", []string{"http://127.0.0.1:9910", "http://127.0.0.1:9911"}, rx.WithHandlers(rHandler))
	provider.Add("/user", []string{"http://127.0.0.1:9910", "http://127.0.0.1:9911"})
	provider.Add("/order", []string{"http://127.0.0.1:9920", "http://127.0.0.1:9921"})
	provider.Add("/book", []string{"http://127.0.0.1:9930", "http://127.0.0.1:9931"})
	provider.Add("/ws", []string{"http://127.0.0.1:9930", "http://127.0.0.1:9931"})

	var s = rx.New()
	s.Load(provider)

	s.Use(func(c *rx.Context) {
		//log.Println("middleware 1")
	})

	s.Use(func(c *rx.Context) {
		//log.Println("middleware 2")
	})

	s.NoRoute(func(c *rx.Context) {
		log.Println("no route:", c.Request.URL.Path)
	})

	s.HandleError(func(c *rx.Context, err error) {
		c.AbortWithJSON(http.StatusBadRequest, "无法访问后端服务")
	})

	var gate = gin.Default()

	//gate.Any("/user/*xx", func(context *gin.Context) {
	//	s.ServeHTTP(context.Writer, context.Request)
	//})
	//gate.Any("/order/*xx", func(context *gin.Context) {
	//	s.ServeHTTP(context.Writer, context.Request)
	//})
	gate.NoRoute(func(context *gin.Context) {
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
