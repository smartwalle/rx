package rx

type Tree struct {
	root *Node
}

func newTree() *Tree {
	var t = &Tree{}
	t.root = newNode("/", 1)
	return t
}

func (this *Tree) Print() {
	this.root.Print()
}

func (this *Tree) Add(path string, handlers ...HandlerFunc) {
	var currentNode = this.root
	if currentNode.key != path {
		var paths = splitPath(path)
		for _, key := range paths {
			var node = currentNode.children[key]
			if node == nil {
				node = newNode(key, currentNode.depth+1)
				currentNode.children[key] = node
			}
			currentNode = node
		}
	}
	currentNode.prepare(path, handlers...)
}

func (this *Tree) find(path string, isRegex bool) (nodes []*Node) {
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

	// 基本上只有 isRegex 为 true 的时候才会执行以下代码
	var queue = make([]*Node, 0, 1)
	queue = append(queue, node)
	// 将 queue 列表中满足条件的 Node 及其满足条件的子 Node 添加到 nodes 列表中
	for len(queue) > 0 {
		var temp []*Node
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
