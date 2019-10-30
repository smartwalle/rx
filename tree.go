package rx

import (
	"sort"
)

type methodTree struct {
	method     string
	root       *treeNode
	numOfNodes int // 拥有的节点数量，只增不减
}

func newMethodTree(method string) *methodTree {
	var t = &methodTree{}
	t.method = method
	t.root = newPathNode("/", 1, t.numOfNodes)
	t.numOfNodes = 1
	return t
}

func (this *methodTree) print() {
	this.root.print()
}

func (this *methodTree) add(path string, handlers HandlerChain) {
	if path == "" {
		return
	}

	var node = this.root
	if node.key != path {
		var paths = splitPath(path)
		for _, key := range paths {
			var child = node.get(key)
			if child == nil {
				this.numOfNodes++
				child = newPathNode(key, node.depth+1, this.numOfNodes)
				node.add(child)
			}
			node = child
		}
	}
	node.prepare(path, handlers)
}

func (this *methodTree) find(path string, isRegex bool, nodes treeNodes) treeNodes {
	if path == "" {
		return nil
	}

	var node = this.root
	var nNodes = nodes

	var paths = splitPath(path)
	if cap(nNodes) < len(paths) {
		nNodes = make(treeNodes, 0, len(paths))
	}

	if node.path == path {
		nNodes = nNodes[:1]
		nNodes[0] = node
		return nNodes
	}

	for _, key := range paths {
		var child = node.get(key)
		if child == nil {
			if isRegex {
				break
			}
			return nNodes
		}

		if child.isValidPath(path) && !isRegex {
			nNodes = nNodes[:1]
			nNodes[0] = child
			return nNodes
		}

		node = child
	}

	if !isRegex {
		return nNodes
	}

	// 只有 isRegex 为 true 的时候才会执行以下代码
	var queue = make([]*treeNode, 0, 1)
	queue = append(queue, node)
	// 将 queue 列表中满足条件的 treeNode 及其满足条件的子 treeNode 添加到 nodes 列表中
	for len(queue) > 0 {
		var temp []*treeNode
		for _, qNode := range queue {
			// 只添加拥有有效路径和正则表达式的节点，以减少后续正则匹配的次数
			if qNode.isValidRegexPath() {
				nNodes = append(nNodes, qNode)
			}

			var children = qNode.children()
			if len(children) > 0 {
				temp = append(temp, children...)
			}
		}
		queue = temp
	}

	// 对 nodes 进行排序
	sort.Sort(nNodes)

	return nNodes
}

func (this *methodTree) clean(path string) {
	if path == "" {
		return
	}

	var node = this.root
	var nodes = make([]*treeNode, 1, 1)
	nodes[0] = node

	if node.path != path {
		// 查询出 path 对应的节点链路及其对应的节点
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
		// 将该节点重置
		node.unprepare()
		// 清理节点链路，移除无效的节点
		var nodeLen = len(nodes)
		for i := nodeLen - 1; i >= 0; i-- {
			var child = nodes[i]
			// 如果该节点是一个有效路径，则终止
			if child.isPath() {
				return
			}
			// 如果该节点没有子节点，则把该节点从其父节点中移除
			if child.numOfChildren() == 0 && i != 0 {
				var parent = nodes[i-1]
				parent.remove(child.key)
			}
		}
	}
}
