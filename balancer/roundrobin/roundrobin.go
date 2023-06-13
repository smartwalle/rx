package roundrobin

import (
	"errors"
	"github.com/smartwalle/rx/balancer"
	"net/http"
	"net/http/httputil"
	"sync"
)

const (
	Name = "round_robin"
)

func New() balancer.Builder {
	return &rrBuilder{}
}

type rrBuilder struct {
}

func (this *rrBuilder) Name() string {
	return Name
}

func (this *rrBuilder) Build(info balancer.BuildInfo) (balancer.Balancer, error) {
	if len(info.Targets) == 0 {
		return nil, errors.New("no targets is available")
	}
	var nTargets = make([]*httputil.ReverseProxy, 0, len(info.Targets))
	for _, value := range info.Targets {
		nTargets = append(nTargets, value)
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
