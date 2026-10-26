package rx

// Router 负责根据 method 和 path 匹配并返回 Route。
type Router interface {
	Match(method string, path string, params Params) (Route, Params, error)
}

type noop struct {
}

func (p *noop) Match(_ string, _ string, _ Params) (Route, Params, error) {
	panic(`router should be specified. 
var router = yourRouter{}
var engine = rx.New()
engine.UseRouter(router)`)
}
