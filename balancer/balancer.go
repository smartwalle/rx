package balancer

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type BuildInfo struct {
	Targets map[*url.URL]*httputil.ReverseProxy
}

type PickResult struct {
	Target *url.URL
	Proxy  *httputil.ReverseProxy
}

type Builder interface {
	Name() string

	Build(info BuildInfo) (Balancer, error)
}

type Balancer interface {
	Pick(req *http.Request) PickResult
}
