// Package pathtree 提供面向路径规则的压缩前缀树。
//
// 规则支持三类片段：静态片段、:name 单段参数和 *name 尾部通配参数，适合在黑名单、白名单、权限规则或 URL 路径规则中保存一批规则，
// 并在查询时返回命中的规则、关联值以及参数捕获结果。
package rx

import (
	"fmt"
	"strings"
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

// Param 是一次路径匹配得到的单个参数。
type Param struct {
	// Key 是规则中声明的参数名，不包含前缀的 : 或 *。
	Key string
	// Value 是输入路径中被该参数捕获到的原始字符串。
	Value string
}

// Params 是按规则中参数出现顺序保存的参数列表。
type Params []Param

// Get 返回指定名称的参数。
func (ps Params) Get(name string) (string, bool) {
	for _, item := range ps {
		if item.Key == name {
			return item.Value, true
		}
	}
	return "", false
}

// ByName 返回指定名称的参数；不存在时返回空字符串。
func (ps Params) ByName(name string) string {
	value, _ := ps.Get(name)
	return value
}

type entry[V any] struct {
	rule  string
	value V
}

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

	// entryIndex 指向 Tree.entries 中的规则和值；-1 表示当前节点不保存规则。
	entryIndex int
}

// Tree 是用于路径规则匹配的压缩前缀树。
//
// Tree 用于保存和匹配路径规则：
//   - 静态路径精确匹配，例如 /users/list；
//   - :name 匹配单个路径段，例如 /users/:id 会匹配 /users/42；
//   - *name 匹配剩余路径，只能出现在规则末尾，并且前一个字节必须是 /；
//   - 规则和待匹配路径都必须以 / 开头；
//   - 每个路径段最多只能包含一个通配符，通配符必须有非空名称；
//   - 同一位置存在静态规则和参数规则时，匹配优先选择静态规则。
//
// Tree 保留规则的插入顺序，Rules 会按这个顺序返回规则；Delete 为了保持压缩结构
// 简单可靠，会删除条目后重建内部规则树。
//
// Tree 支持多个 goroutine 并发读取。只要存在 goroutine 会写入同一个 Tree，
// 调用方就需要自行加锁保护读写和写写并发。
type Tree[V any] struct {
	root    *node
	entries []entry[V]
	index   map[string]int
}

// NewTree 创建一个空 Tree。
func NewTree[V any]() *Tree[V] {
	return &Tree[V]{
		root:    newNode(nodeRoot),
		entries: make([]entry[V], 0),
		index:   make(map[string]int),
	}
}

// Len 返回当前规则数量。
//
// Len 对 nil Tree 返回 0，方便调用方在可选 Tree 场景中直接使用。
func (t *Tree[V]) Len() int {
	if t == nil {
		return 0
	}
	return len(t.entries)
}

// Get 根据 rule 获取保存的 value。
//
// Get 只检查规则本身是否已保存，不会把 rule 当作输入执行规则匹配。
func (t *Tree[V]) Get(rule string) (V, bool) {
	var zero V
	if t == nil || t.index == nil {
		return zero, false
	}
	index, ok := t.index[rule]
	if !ok {
		return zero, false
	}
	return t.entries[index].value, true
}

// Has 判断 rule 是否已保存。
//
// Has 只检查规则本身是否存在，不会把 rule 当作输入执行规则匹配。
func (t *Tree[V]) Has(rule string) bool {
	if t == nil || t.index == nil {
		return false
	}
	_, ok := t.index[rule]
	return ok
}

// Put 写入 rule 和 value。
//
// 如果 rule 不存在，Put 会插入新规则，并返回 value 的零值、false 和 nil。
// 如果 rule 已存在，Put 会覆盖旧 value，并返回旧 value、true 和 nil。
//
// rule 必须以 / 开头，通配符也必须符合路径规则语义。非法规则或与已有通配规则冲突时，
// Put 会返回错误，且不会写入规则索引。
func (t *Tree[V]) Put(rule string, value V) (V, bool, error) {
	var zero V
	if t == nil {
		return zero, false, nil
	}
	if !validPath(rule) {
		return zero, false, fmt.Errorf("path must begin with '/'")
	}
	t.ensure()
	if index, ok := t.index[rule]; ok {
		old := t.entries[index].value
		t.entries[index].value = value
		return old, true, nil
	}

	if err := t.root.addRule(rule, len(t.entries)); err != nil {
		return zero, false, err
	}
	t.entries = append(t.entries, entry[V]{rule: rule, value: value})
	t.index[rule] = len(t.entries) - 1
	return zero, false, nil
}

// Set 写入 rule 和 value，并忽略旧值。
//
// Set 适合调用方只关心最终写入结果、不关心 rule 原来是否存在的场景。
// 如果需要知道本次写入是新增还是覆盖，或者需要拿到旧 value，应使用 Put。
// 如果 rule 非法或与已有通配规则冲突，Set 会返回错误。
func (t *Tree[V]) Set(rule string, value V) error {
	_, _, err := t.Put(rule, value)
	return err
}

// Delete 删除指定 rule。
//
// 如果 rule 存在，Delete 会删除该规则，并返回被删除的 value 和 true。
// 如果 rule 不存在，Delete 返回 value 的零值和 false。
func (t *Tree[V]) Delete(rule string) (V, bool) {
	var zero V
	if t == nil || t.index == nil {
		return zero, false
	}
	index, ok := t.index[rule]
	if !ok {
		return zero, false
	}

	old := t.entries[index]
	entries := make([]entry[V], 0, len(t.entries)-1)
	entries = append(entries, t.entries[:index]...)
	entries = append(entries, t.entries[index+1:]...)
	root, table, err := build(entries)
	if err != nil {
		return zero, false
	}
	t.entries = entries
	t.root = root
	t.index = table
	return old.value, true
}

// Clear 清空所有规则。
//
// Clear 会直接丢弃内部索引和规则树，让旧数据交给 GC 回收。
func (t *Tree[V]) Clear() {
	if t == nil {
		return
	}
	t.root = nil
	t.entries = nil
	t.index = nil
}

// Rules 返回按插入顺序排列的规则。
//
// 返回的切片是副本，调用方修改它不会影响 Tree。
func (t *Tree[V]) Rules() []string {
	if t == nil || len(t.entries) == 0 {
		return nil
	}
	rules := make([]string, len(t.entries))
	for i, item := range t.entries {
		rules[i] = item.rule
	}
	return rules
}

// Match 判断 path 是否命中已保存的规则，并返回参数。
//
// 如果匹配成功，返回命中的完整规则、对应 value、参数列表和 true。
// 如果没有命中，返回空规则、value 零值、nil Params 和 false。
//
// Params 中会保存 :name 和 *name 捕获到的值，顺序与规则中的通配符出现顺序一致。
// :name 不会跨过 /；*name 会捕获规则剩余部分，包括开头的 /。
func (t *Tree[V]) Match(path string) (string, V, Params, bool) {
	var zero V
	if t == nil || t.root == nil {
		return "", zero, nil, false
	}
	params := make(Params, 0)
	rule, value, params, ok := matchNode(t.root, t.entries, path, params, true)
	if !ok {
		return "", zero, nil, false
	}
	return rule, value, params, true
}

func (t *Tree[V]) ensure() {
	if t.root == nil {
		t.root = newNode(nodeRoot)
	}
	if t.index == nil {
		t.index = make(map[string]int)
	}
}

func build[V any](entries []entry[V]) (*node, map[string]int, error) {
	root := newNode(nodeRoot)
	table := make(map[string]int, len(entries))
	for index, item := range entries {
		table[item.rule] = index
		if err := root.addRule(item.rule, index); err != nil {
			return nil, nil, err
		}
	}
	return root, table, nil
}

func (t *Tree[V]) rebuild() error {
	root, table, err := build(t.entries)
	if err != nil {
		return err
	}
	t.root = root
	t.index = table
	return nil
}

func (n *node) addRule(path string, entryIndex int) error {
	current := n
	fullRule := path
	current.priority++
	if current.path == "" && len(current.children) == 0 {
		if err := current.insertChild(path, fullRule, entryIndex); err != nil {
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
				path:       current.path[common:],
				indices:    current.indices,
				wildChild:  current.wildChild,
				nType:      nodeStatic,
				priority:   current.priority - 1,
				children:   current.children,
				entryIndex: current.entryIndex,
			}
			current.children = []*node{child}
			current.indices = byteString(current.path[common])
			current.path = path[:common]
			current.entryIndex = -1
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
				return wildcardConflictError(path, fullRule, current)
			}

			return current.insertChild(path, fullRule, entryIndex)
		}

		current.entryIndex = entryIndex
		return nil
	}
}

func (n *node) insertChild(path string, fullRule string, entryIndex int) error {
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
			return fmt.Errorf("only one wildcard per path segment is allowed, has: %q in path %q", wildcard, fullRule)
		}
		if len(wildcard) < 2 {
			return fmt.Errorf("wildcards must be named with a non-empty name in path %q", fullRule)
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

			current.entryIndex = entryIndex
			return nil
		}

		// catch-all 由两层节点表示：第一层作为通配入口，第二层保存参数名和 value。
		// 这样能和静态 "/" 子节点保持相同的索引路径。
		if index+len(wildcard) != len(path) {
			return fmt.Errorf("catch-all rules are only allowed at the end of the path in path %q", fullRule)
		}
		if index == 0 || path[index-1] != '/' {
			return fmt.Errorf("no / before catch-all in path %q", fullRule)
		}
		if len(current.path) > 0 && current.path[len(current.path)-1] == '/' {
			return fmt.Errorf("catch-all wildcard %q in new rule %q conflicts with existing path segment in existing prefix %q", path, fullRule, current.path)
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
		child.entryIndex = entryIndex
		current.children = []*node{child}
		return nil
	}

	current.path = path
	current.entryIndex = entryIndex
	return nil
}

func matchNode[V any](n *node, entries []entry[V], path string, params Params, collectParams bool) (string, V, Params, bool) {
	var zero V
	current := n
	if len(path) < len(current.path) || path[:len(current.path)] != current.path {
		return "", zero, params, false
	}
	path = path[len(current.path):]
	if current.nType == nodeCatchAll && current.wildChild && len(current.children) > 0 {
		leaf := current.children[0]
		if collectParams {
			params = append(params, Param{Key: leaf.path[2:], Value: path})
		}
		return matchEntry(entries, leaf.entryIndex, params)
	}
	if path == "" {
		if current.entryIndex >= 0 {
			return matchEntry(entries, current.entryIndex, params)
		}
		return "", zero, params, false
	}

	if len(current.indices) > 0 {
		c := path[0]
		for i := 0; i < len(current.indices); i++ {
			if c == current.indices[i] {
				if rule, value, gotParams, ok := matchNode(current.children[i], entries, path, params, collectParams); ok {
					return rule, value, gotParams, true
				}
				break
			}
		}
	}

	if !current.wildChild || len(current.children) == 0 {
		return "", zero, params, false
	}

	child := current.children[len(current.children)-1]
	switch child.nType {
	case nodeParam:
		end := 0
		for end < len(path) && path[end] != '/' {
			end++
		}
		if collectParams {
			params = append(params, Param{Key: child.path[1:], Value: path[:end]})
		}
		if end < len(path) {
			if len(child.children) == 0 {
				return "", zero, params, false
			}
			return matchNode(child.children[0], entries, path[end:], params, collectParams)
		}
		if child.entryIndex >= 0 {
			return matchEntry(entries, child.entryIndex, params)
		}
		return "", zero, params, false
	case nodeCatchAll:
		if len(child.children) == 0 {
			return "", zero, params, false
		}
		leaf := child.children[0]
		if collectParams {
			params = append(params, Param{Key: leaf.path[2:], Value: path})
		}
		return matchEntry(entries, leaf.entryIndex, params)
	default:
		return "", zero, params, false
	}
}

func matchEntry[V any](entries []entry[V], index int, params Params) (string, V, Params, bool) {
	var zero V
	if index < 0 || index >= len(entries) {
		return "", zero, params, false
	}
	item := entries[index]
	return item.rule, item.value, params, true
}

func newNode(typ nodeType) *node {
	return &node{nType: typ, entryIndex: -1}
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

func wildcardConflictError(path string, fullRule string, n *node) error {
	pathSeg := path
	if n.nType != nodeCatchAll {
		pathSeg, _, _ = strings.Cut(pathSeg, "/")
	}
	prefix := fullRule[:strings.Index(fullRule, pathSeg)] + n.path
	return fmt.Errorf("%q in new rule %q conflicts with existing wildcard %q in existing prefix %q", pathSeg, fullRule, n.path, prefix)
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
