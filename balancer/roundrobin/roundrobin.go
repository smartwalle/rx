package roundrobin

import (
	"errors"
	"github.com/smartwalle/rx/balancer"
	"net/http"
	"sync/atomic"
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
	var nTargets = make([]balancer.PickResult, 0, len(info.Targets))
	for key, value := range info.Targets {
		nTargets = append(nTargets, balancer.PickResult{Target: key, Proxy: value})
	}
	return &rrBalancer{targets: nTargets, size: len(nTargets), offset: 0}, nil
}

type rrBalancer struct {
	targets []balancer.PickResult
	size    int
	offset  uint32
}

func (this *rrBalancer) Pick(req *http.Request) (balancer.PickResult, error) {
	if this.size == 0 {
		return balancer.PickResult{}, errors.New("no targets is available")
	}
	var index = int(atomic.AddUint32(&this.offset, 1)-1) % this.size
	target := this.targets[index]
	return target, nil
}
