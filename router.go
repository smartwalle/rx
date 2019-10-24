package rx

type Router struct {
	trees map[string]*Tree
}

func NewRouter() *Router {
	var r = &Router{}
	r.trees = make(map[string]*Tree)
	return r
}
