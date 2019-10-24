package rx

import (
	"fmt"
	"net/http"
)

type Node struct {
	name     string
	path     string
	depth    int
	children map[string]*Node
	handlers []http.HandlerFunc
}

func newNode(name string, depth int) *Node {
	var n = &Node{}
	n.name = name
	n.depth = depth
	n.children = make(map[string]*Node)
	return n
}

func (this *Node) String() string {
	return fmt.Sprintf("Name:%s  Path:%s", this.name, this.path)
}

func (this *Node) Print() {
	for i := 0; i < this.depth; i++ {
		fmt.Print("-")
	}
	fmt.Println(this.String())
	for _, c := range this.children {
		c.Print()
	}
}
