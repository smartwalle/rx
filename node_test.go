package rx

import (
	"fmt"
	"testing"
)

func TestNode_Prepare(t *testing.T) {

	var h = HandlerChain{}
	h = append(h, func(c *Context) {})

	var n = newPathNode("h1", 1, 1)
	n.prepare("/path1/:p1/:p2/{p3:([\\w]+)}", h)

	fmt.Println(n.match("/path1/v1/v2/v3"))
	fmt.Println(n.match("/path1/wv1/wv2/wv3"))
	fmt.Println(n.match("/path1/wv1/wv2"))
	fmt.Println(n.match("/path1/wv1/wv2/wv3/wv4"))
}
