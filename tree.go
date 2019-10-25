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

	var currentNode = this.root
	if currentNode.key != path {
		var paths = splitPath(path)
		for _, key := range paths {
			var node = currentNode.children[key]
			if node == nil {
				node = newPathNode(key, currentNode.depth+1)
				currentNode.children[key] = node
			}
			currentNode = node
		}
	}
	currentNode.prepare(path, handlers...)
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
		var child = node.children[key]
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
