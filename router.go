package rx

import "net/http"

// Router 负责根据 method 和 path 匹配并返回 Route。
type Router interface {
	Match(method string, path string, params Params) (Route, Params, error)
}

type noop struct {
}

func (p *noop) Match(_ string, path string, params Params) (Route, Params, error) {
	return NewRoute(path, func(c *Context) {
		c.String(http.StatusBadRequest, `router should be specified:

var r = router.New()
r.GET("/health", func(c *rx.Context) {
    c.JSON(http.StatusOK, "ok")
})

var engine = rx.New()
engine.UseRouter(r)
...
...
`)
	}), params, nil
}
