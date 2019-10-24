package rx

import "net/http"

type Node struct {
	name     string
	path     string
	depth    int
	children map[string]*Node
	handlers []http.HandlerFunc
}

func NewNode(name string, depth int) *Node {
	var n = &Node{}
	n.name = name
	n.depth = depth
	return n
}
