package roundrobin

import (
	"errors"
	"github.com/smartwalle/rx/balancer"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type Builder struct {
}

func (this *Builder) Name() string {
	return "round-robin"
}

func (this *Builder) Build(targets []*url.URL) (balancer.Balancer, error) {
	if len(targets) == 0 {
		return nil, errors.New("no targets is available")
	}

	var nTargets = make([]*httputil.ReverseProxy, 0, len(targets))
	for _, target := range targets {
		nTargets = append(nTargets, httputil.NewSingleHostReverseProxy(target))
	}
	return &rrBalancer{targets: nTargets, next: 0}, nil
}

type rrBalancer struct {
	targets []*httputil.ReverseProxy
	next    int
	mu      sync.Mutex
}

func (this *rrBalancer) Pick(req *http.Request) (*httputil.ReverseProxy, error) {
	this.mu.Lock()
	target := this.targets[this.next]
	this.next = (this.next + 1) % len(this.targets)
	this.mu.Unlock()
	return target, nil
}
