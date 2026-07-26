package main

import (
	"log"
	"net/http"
	"regexp"

	"github.com/smartwalle/rx"
)

type gateHTTPRouter struct {
	routes []gateHTTPRoute
}

type gateHTTPRoute struct {
	regexp *regexp.Regexp
	route  rx.Route
}

func (r *gateHTTPRouter) Add(pattern string, targets []string, opts ...rx.Option) error {
	route, err := rx.NewProxyRoute(pattern, targets, opts...)
	if err != nil {
		return err
	}
	nRegexp, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	r.routes = append(r.routes, gateHTTPRoute{regexp: nRegexp, route: route})
	return nil
}

func (r *gateHTTPRouter) Match(req *http.Request) (rx.Route, error) {
	for _, route := range r.routes {
		if route.regexp.MatchString(req.URL.Path) {
			return route.route, nil
		}
	}
	return nil, nil
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	var provider = &gateHTTPRouter{}
	provider.Add("/user", []string{"http://127.0.0.1:9910", "http://127.0.0.1:9911"})
	provider.Add("/order", []string{"http://127.0.0.1:9920", "http://127.0.0.1:9921"})
	provider.Add("/ws", []string{"http://127.0.0.1:9930", "http://127.0.0.1:9931"})

	var s = rx.New()
	s.UseRouter(provider)

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
