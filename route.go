package rx

// Route 表示 Router 已经匹配到的处理目标。
type Route interface {
	Path() string
	Handlers() HandlersChain
}

type localRoute struct {
	path     string
	handlers HandlersChain
}

// NewRoute 创建只执行本地处理函数的路由。
func NewRoute(path string, handlers ...HandlerFunc) Route {
	var route = &localRoute{}
	route.path = path
	route.handlers = handlers
	return route
}

func (route *localRoute) Path() string {
	return route.path
}

func (route *localRoute) Handlers() HandlersChain {
	return route.handlers
}

func CombineHandlers(base HandlersChain, handlers ...HandlerFunc) HandlersChain {
	if len(base) == 0 && len(handlers) == 0 {
		return nil
	}
	combined := make(HandlersChain, 0, len(base)+len(handlers))
	combined = append(combined, base...)
	combined = append(combined, handlers...)
	return combined
}
