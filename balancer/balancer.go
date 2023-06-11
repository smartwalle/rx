package balancer

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Builder interface {
	Name() string

	Build(targets []*url.URL) (Balancer, error)
}

type Balancer interface {
	Pick(req *http.Request) (*httputil.ReverseProxy, error)
}
