package rx

type Tree struct {
	root *Node
}

func NewTree() *Tree {
	var t = &Tree{}
	t.root = NewNode("/", 1)
	return t
}
