package rx

import (
	"github.com/smartwalle/rx/balancer"
	"github.com/smartwalle/rx/balancer/roundrobin"
	"strings"
)

var balancers = make(map[string]balancer.Builder)

func init() {
	RegisterBalancer(roundrobin.New())
}

func RegisterBalancer(b balancer.Builder) {
	if b != nil {
		balancers[strings.ToLower(b.Name())] = b
	}
}

func GetBalancer(name string) balancer.Builder {
	if b, ok := balancers[strings.ToLower(name)]; ok {
		return b
	}
	return nil
}

func DefaultBalancer() balancer.Builder {
	return balancers[roundrobin.Name]
}
