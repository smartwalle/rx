package rx

import "net/http"

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

func (this *Tree) Add(path string, handlers ...http.HandlerFunc) {
	var currentNode = this.root
	if currentNode.name != path {
		var paths = splitPath(path)
		for _, name := range paths {
			var node = currentNode.children[name]
			if node == nil {
				node = newNode(name, currentNode.depth+1)
				currentNode.children[name] = node
			}
			currentNode = node
		}
	}
	currentNode.path = path
	currentNode.handlers = handlers
}

func (this *Tree) Find(path string, isRegex bool) (nodes []*Node) {
	var node = this.root

	if node.path == path {
		nodes = append(nodes, node)
		return nodes
	}

	var paths = splitPath(path)
	for _, name := range paths {
		var child = node.children[name]
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

	return nil
}

func (this *Tree) FindOne(path string) *Node {
	var node = this.root

	if node.path == path {
		return node
	}

	var paths = splitPath(path)
	for _, name := range paths {
		var child = node.children[name]
		if child == nil {
			return nil
		}

		node = child

		if child.path == path {
			return node
		}
	}

	return nil
}
