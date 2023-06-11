package rx

import (
	"github.com/smartwalle/rx/balancer"
)

type Route struct {
	handlers HandlersChain
	balancer balancer.Balancer
}
