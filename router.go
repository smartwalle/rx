package rx

import (
	"net/http"
)

// Router 负责根据请求匹配并返回 Route。
type Router interface {
	Match(req *http.Request) (Route, error)
}

type nonRouter struct {
}

func (p *nonRouter) Match(req *http.Request) (Route, error) {
	panic(`router should be specified. 
var router = yourRouter{}
var engine = rx.New()
engine.Load(router)`)
}
