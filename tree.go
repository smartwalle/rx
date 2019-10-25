package rx

type methodTree struct {
	root *pathNode
}

func newMethodTree() *methodTree {
	var t = &methodTree{}
	t.root = newPathNode("/", 1)
	return t
}

func (this *methodTree) Print() {
	this.root.Print()
}

func (this *methodTree) add(path string, handlers ...HandlerFunc) {
	if path == "" {
		return
	}

	var node = this.root
	if node.key != path {
		var paths = splitPath(path)
		for _, key := range paths {
			var child = node.get(key)
			if child == nil {
				child = newPathNode(key, node.depth+1)
				node.add(child)
			}
			node = child
		}
	}
	node.prepare(path, handlers...)
}

func (this *methodTree) find(path string, isRegex bool) (nodes []*pathNode) {
	if path == "" {
		return nil
	}

	var node = this.root

	if node.path == path {
		nodes = append(nodes, node)
		return nodes
	}

	var paths = splitPath(path)
	for _, key := range paths {
		var child = node.get(key)
		if child == nil {
			if isRegex {
				break
			}
			return nil
		}

		if child.path == path && !isRegex {
			nodes = append(nodes, child)
			return nodes
		}

		node = child
	}

	if !isRegex {
		return nil
	}

	// 只有 isRegex 为 true 的时候才会执行以下代码
	var queue = make([]*pathNode, 0, 1)
	queue = append(queue, node)
	// 将 queue 列表中满足条件的 pathNode 及其满足条件的子 pathNode 添加到 nodes 列表中
	for len(queue) > 0 {
		var temp []*pathNode
		for _, qNode := range queue {
			if qNode.isPath {
				nodes = append(nodes, qNode)
			}
			for _, child := range qNode.children {
				temp = append(temp, child)
			}
		}
		queue = temp
	}

	return nodes
}

func (this *methodTree) clean(path string) {
	if path == "" {
		return
	}

	var node = this.root
	var nodes = make([]*pathNode, 1, 1)
	nodes[0] = node

	if node.path != path {
		var paths = splitPath(path)
		for _, key := range paths {
			var child = node.get(key)
			if child == nil {
				return
			}

			node = child
			nodes = append(nodes, child)

			if child.path == path {
				break
			}
		}
	}

	if node != nil {
		node.reset()
		var nodeLen = len(nodes)
		for i := nodeLen - 1; i >= 0; i-- {
			var child = nodes[i]
			if child.isPath {
				return
			}
			if len(child.children) == 0 && i != 0 {
				var parent = nodes[i-1]
				parent.remove(child.key)
			}
		}
	}
}
