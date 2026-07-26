package router

import (
	"fmt"
	"strings"

	"github.com/smartwalle/rx"
)

type nodeType uint8

const (
	// nodeStatic 表示普通静态节点，path 按字面量匹配。
	nodeStatic nodeType = iota
	// nodeRoot 表示树的根节点。
	nodeRoot
	// nodeParam 表示 :name 参数节点，只匹配单个路径段。
	nodeParam
	// nodeCatchAll 表示 *name 通配节点，只能位于规则末尾。
	nodeCatchAll
)

type node struct {
	// path 保存从父节点到当前节点需要消费的压缩路径片段。
	path string
	// indices 保存所有静态子节点 path 的首字节，顺序与 children 中的静态子节点一致。
	indices string
	// wildChild 表示 children 的最后一个节点是 nodeParam 或 nodeCatchAll 通配节点。
	wildChild bool
	nType     nodeType
	// priority 记录节点热度，插入相同前缀时会把高优先级静态子节点前移。
	priority uint32
	children []*node

	fullPath string
	route    rx.Route
}

// tree 是基于压缩前缀树的路径路由器。
//
// tree 用于保存和匹配路径：
//   - 静态路径精确匹配，例如 /users/list；
//   - :name 匹配单个路径段，例如 /users/:id 会匹配 /users/42；
//   - *name 匹配剩余路径，只能出现在规则末尾，并且前一个字节必须是 /；
//   - 注册路径和待匹配路径都必须以 / 开头；
//   - 每个路径段最多只能包含一个通配符，通配符必须有非空名称；
//   - 同一位置存在静态规则和参数规则时，匹配优先选择静态规则。
//
// tree 支持多个 goroutine 并发读取。只要存在 goroutine 会写入同一个 tree，
// 调用方就需要自行加锁保护读写和写写并发。
type tree struct {
	root *node
}

func newTree() *tree {
	return &tree{
		root: newNode(nodeRoot),
	}
}

// add 注册 Route。
//
// Route.Path 必须以 / 开头，通配符也必须符合路径规则语义。非法规则或与已有通配规则冲突时，
// add 会返回错误。
func (t *tree) add(route rx.Route) error {
	if t == nil {
		return nil
	}
	if route == nil {
		return nil
	}
	path := route.Path()
	if !validPath(path) {
		return fmt.Errorf("path must begin with '/'")
	}
	t.ensure()
	if mNode, _, ok := matchNode(t.root, path, nil, false); ok && mNode.fullPath == path {
		return fmt.Errorf("route already exists for path %q", path)
	}

	return t.root.addRoute(path, route)
}

func (t *tree) match(path string, params rx.Params) (rx.Route, rx.Params, error) {
	params = params[:0]
	if t == nil {
		return nil, params, nil
	}
	if t.root == nil {
		return nil, params, nil
	}
	mNode, params, ok := matchNode(t.root, path, params, true)
	if !ok {
		return nil, params[:0], nil
	}
	return mNode.route, params, nil
}

func (t *tree) ensure() {
	if t.root == nil {
		t.root = newNode(nodeRoot)
	}
}

func (n *node) addRoute(path string, route rx.Route) error {
	current := n
	fullPath := path
	current.priority++
	if current.path == "" && len(current.children) == 0 {
		if err := current.insertChild(path, fullPath, route); err != nil {
			return err
		}
		current.nType = nodeRoot
		return nil
	}

walk:
	for {
		common := longestCommonPrefix(path, current.path)
		if common < len(current.path) {
			// 当前压缩边只与新规则共享一部分前缀，需要拆成公共父节点和旧子节点。
			child := &node{
				path:      current.path[common:],
				indices:   current.indices,
				wildChild: current.wildChild,
				nType:     nodeStatic,
				priority:  current.priority - 1,
				children:  current.children,
				fullPath:  current.fullPath,
				route:     current.route,
			}
			current.children = []*node{child}
			current.indices = byteString(current.path[common])
			current.path = path[:common]
			current.fullPath = ""
			current.route = nil
			current.wildChild = false
		}

		if common < len(path) {
			path = path[common:]
			c := path[0]

			if current.nType == nodeParam && c == '/' && len(current.children) == 1 {
				current = current.children[0]
				current.priority++
				continue walk
			}

			for i := 0; i < len(current.indices); i++ {
				if c == current.indices[i] {
					i = current.incrementChildPriority(i)
					current = current.children[i]
					continue walk
				}
			}

			if c != ':' && c != '*' && current.nType != nodeCatchAll {
				current.indices += byteString(c)
				child := newNode(nodeStatic)
				current.addChild(child)
				current.incrementChildPriority(len(current.indices) - 1)
				current = child
			} else if current.wildChild {
				current = current.children[len(current.children)-1]
				current.priority++
				if len(path) >= len(current.path) && current.path == path[:len(current.path)] &&
					current.nType != nodeCatchAll &&
					(len(current.path) >= len(path) || path[len(current.path)] == '/') {
					continue walk
				}
				return wildcardConflictError(path, fullPath, current)
			}

			return current.insertChild(path, fullPath, route)
		}

		if current.route != nil {
			return fmt.Errorf("route already exists for path %q", fullPath)
		}
		current.fullPath = fullPath
		current.route = route
		return nil
	}
}

func (n *node) insertChild(path string, fullPath string, route rx.Route) error {
	current := n
	for {
		wildcard, index, valid, err := findWildcard(path)
		if err != nil {
			return err
		}
		if index < 0 {
			break
		}
		if !valid {
			return fmt.Errorf("only one wildcard per path segment is allowed, has: %q in path %q", wildcard, fullPath)
		}
		if len(wildcard) < 2 {
			return fmt.Errorf("wildcards must be named with a non-empty name in path %q", fullPath)
		}

		if wildcard[0] == ':' {
			// 参数节点独占当前路径段，后续静态子路径会挂在参数节点下面继续匹配。
			if index > 0 {
				current.path = path[:index]
				path = path[index:]
			}

			child := newNode(nodeParam)
			child.path = wildcard
			current.addChild(child)
			current.wildChild = true
			current = child
			current.priority++

			if len(wildcard) < len(path) {
				path = path[len(wildcard):]
				child = newNode(nodeStatic)
				child.priority = 1
				current.addChild(child)
				current = child
				continue
			}

			if current.route != nil {
				return fmt.Errorf("route already exists for path %q", fullPath)
			}
			current.fullPath = fullPath
			current.route = route
			return nil
		}

		// catch-all 由两层节点表示：第一层作为通配入口，第二层保存参数名和 Route。
		// 这样能和静态 "/" 子节点保持相同的索引路径。
		if index+len(wildcard) != len(path) {
			return fmt.Errorf("catch-all wildcards are only allowed at the end of path %q", fullPath)
		}
		if index == 0 || path[index-1] != '/' {
			return fmt.Errorf("no / before catch-all in path %q", fullPath)
		}
		if len(current.path) > 0 && current.path[len(current.path)-1] == '/' {
			return fmt.Errorf("catch-all wildcard %q in path %q conflicts with existing path segment in existing prefix %q", path, fullPath, current.path)
		}

		current.path = path[:index-1]
		child := newNode(nodeCatchAll)
		child.wildChild = true
		current.addChild(child)
		current.indices = "/"
		current = child
		current.priority++

		child = newNode(nodeCatchAll)
		child.path = path[index-1:]
		child.priority = 1
		child.fullPath = fullPath
		child.route = route
		current.children = []*node{child}
		return nil
	}

	current.path = path
	current.fullPath = fullPath
	current.route = route
	return nil
}

func matchNode(n *node, path string, params rx.Params, collectParams bool) (*node, rx.Params, bool) {
	current := n
	if len(path) < len(current.path) || path[:len(current.path)] != current.path {
		return nil, params, false
	}
	path = path[len(current.path):]
	if current.nType == nodeCatchAll && current.wildChild && len(current.children) > 0 {
		leaf := current.children[0]
		if collectParams {
			params = append(params, rx.Param{Key: leaf.path[2:], Value: path})
		}
		return matchRoute(leaf, params)
	}
	if path == "" {
		if current.route != nil {
			return current, params, true
		}
		return nil, params, false
	}

	if len(current.indices) > 0 {
		c := path[0]
		for i := 0; i < len(current.indices); i++ {
			if c == current.indices[i] {
				if mNode, gotParams, ok := matchNode(current.children[i], path, params, collectParams); ok {
					return mNode, gotParams, true
				}
				break
			}
		}
	}

	if !current.wildChild || len(current.children) == 0 {
		return nil, params, false
	}

	child := current.children[len(current.children)-1]
	switch child.nType {
	case nodeParam:
		end := 0
		for end < len(path) && path[end] != '/' {
			end++
		}
		if collectParams {
			params = append(params, rx.Param{Key: child.path[1:], Value: path[:end]})
		}
		if end < len(path) {
			if len(child.children) == 0 {
				return nil, params, false
			}
			return matchNode(child.children[0], path[end:], params, collectParams)
		}
		if child.route != nil {
			return child, params, true
		}
		return nil, params, false
	case nodeCatchAll:
		if len(child.children) == 0 {
			return nil, params, false
		}
		leaf := child.children[0]
		if collectParams {
			params = append(params, rx.Param{Key: leaf.path[2:], Value: path})
		}
		return matchRoute(leaf, params)
	default:
		return nil, params, false
	}
}

func matchRoute(n *node, params rx.Params) (*node, rx.Params, bool) {
	if n == nil || n.route == nil {
		return nil, params, false
	}
	return n, params, true
}

func newNode(typ nodeType) *node {
	return &node{nType: typ}
}

func (n *node) addChild(child *node) {
	// 通配子节点必须始终位于 children 末尾；匹配时会先尝试静态子节点，再兜底通配。
	if n.wildChild && len(n.children) > 0 {
		wildcard := n.children[len(n.children)-1]
		n.children = append(n.children[:len(n.children)-1], child, wildcard)
		return
	}
	n.children = append(n.children, child)
}

func (n *node) incrementChildPriority(position int) int {
	child := n.children[position]
	child.priority++
	newPosition := position
	for newPosition > 0 && n.children[newPosition-1].priority < child.priority {
		n.children[newPosition] = n.children[newPosition-1]
		newPosition--
	}
	if newPosition == position {
		return position
	}
	n.children[newPosition] = child
	n.indices = n.indices[:newPosition] + n.indices[position:position+1] + n.indices[newPosition:position] + n.indices[position+1:]
	return newPosition
}

func findWildcard(path string) (wildcard string, index int, valid bool, err error) {
	escapeColon := false
	for start := 0; start < len(path); start++ {
		c := path[start]
		if escapeColon {
			escapeColon = false
			if c == ':' {
				continue
			}
			return "", -1, false, fmt.Errorf("invalid escape string in path %q", path)
		}
		if c == '\\' {
			escapeColon = true
			continue
		}
		if c != ':' && c != '*' {
			continue
		}

		valid = true
		for end := start + 1; end < len(path); end++ {
			switch path[end] {
			case '/':
				return path[start:end], start, valid, nil
			case ':', '*':
				valid = false
			}
		}
		return path[start:], start, valid, nil
	}
	return "", -1, false, nil
}

func wildcardConflictError(path string, fullPath string, n *node) error {
	pathSeg := path
	if n.nType != nodeCatchAll {
		pathSeg, _, _ = strings.Cut(pathSeg, "/")
	}
	prefix := fullPath[:strings.Index(fullPath, pathSeg)] + n.path
	return fmt.Errorf("%q in path %q conflicts with existing wildcard %q in existing prefix %q", pathSeg, fullPath, n.path, prefix)
}

func longestCommonPrefix(a string, b string) int {
	limit := len(a)
	if len(b) < limit {
		limit = len(b)
	}
	for i := 0; i < limit; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return limit
}

func validPath(path string) bool {
	return len(path) > 0 && path[0] == '/'
}

func byteString(value byte) string {
	return string([]byte{value})
}
