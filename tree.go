package rx

type methodTree struct {
	method string
	root   *node
}

type methodTrees []*methodTree

func (this methodTrees) get(method string) *methodTree {
	for _, tree := range this {
		if tree.method == method {
			return tree
		}
	}
	return nil
}

func newMethodTree(method string) *methodTree {
	var t = &methodTree{}
	t.method = method
	t.root = &node{}
	return t
}
