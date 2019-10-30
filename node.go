package rx

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

const (
	wildcard1 = `([^\s/]+)`
	wildcard2 = `([\S]+)`
)

type treeNodeChain []*treeNode

func (this treeNodeChain) Len() int {
	return len(this)
}

func (this treeNodeChain) Less(i, j int) bool {
	return this[i].priority < this[j].priority
}

func (this treeNodeChain) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

type treeNode struct {
	mu       sync.RWMutex
	key      string               // 标识
	depth    int                  // 深度
	priority int                  // 节点的优先级，按照节点添加添加的顺序递增，值越小优先级越高，正则匹配的时候，将按照这个顺序进行匹配
	subNodes map[string]*treeNode // 子节点

	path     string       // 对应的路径
	handlers HandlerChain // 对应的 handler 列表

	regex      *regexp.Regexp // path 对应的正则表达式
	paramNames []string       // path 中对应的参数名称列表
}

func newPathNode(key string, depth, priority int) *treeNode {
	var n = &treeNode{}
	n.key = key
	n.depth = depth
	n.priority = priority
	n.subNodes = make(map[string]*treeNode)
	return n
}

func (this *treeNode) numOfChildren() int {
	this.mu.RLock()
	var l = len(this.subNodes)
	this.mu.RUnlock()
	return l
}

func (this *treeNode) children() []*treeNode {
	this.mu.RLock()
	var ns = make([]*treeNode, 0, len(this.subNodes))
	for _, n := range this.subNodes {
		ns = append(ns, n)
	}
	this.mu.RUnlock()
	return ns
}

func (this *treeNode) add(node *treeNode) {
	if node == nil {
		return
	}
	this.mu.Lock()
	this.subNodes[node.key] = node
	this.mu.Unlock()
}

func (this *treeNode) get(key string) *treeNode {
	this.mu.RLock()
	var n = this.subNodes[key]
	this.mu.RUnlock()
	return n
}

func (this *treeNode) remove(key string) {
	this.mu.Lock()
	delete(this.subNodes, key)
	this.mu.Unlock()
}

// prepare 主要对 path 进行预处理，检测该 path 是否包含正则表达式，
// 如果包含正则表达式，则编译成正则表达式对象缓存起来，并提取出相应的参数列表。
func (this *treeNode) prepare(path string, handlers HandlerChain) {
	this.path = path
	this.handlers = handlers

	var paths = splitPath(path)
	var pattern = ""
	var paramsNames = make([]string, 0, len(paths))
	var isRegex = false

	for _, p := range paths {
		if p == "" {
			continue
		}

		var strLen = len(p)
		var lastChar = lastChar(p)
		var firstChar = firstChar(p)

		if firstChar == ':' {
			var name = p[1:strLen]
			pattern = pattern + "/" + wildcard1
			paramsNames = append(paramsNames, name)
			isRegex = true
		} else if firstChar == '*' {
			var name = p[1:strLen]
			pattern = pattern + "/" + wildcard2
			paramsNames = append(paramsNames, name)
			isRegex = true
		} else if firstChar == '{' && lastChar == '}' {
			var subStrList = strings.Split(p[1:strLen-1], ":")
			paramsNames = append(paramsNames, subStrList[0])
			pattern = pattern + "/" + subStrList[1]
			isRegex = true
		} else {
			pattern = pattern + "/" + p
		}
	}

	if isRegex {
		this.regex = regexp.MustCompile(pattern)
		this.paramNames = paramsNames
	}
}

func (this *treeNode) unprepare() {
	this.path = ""
	this.handlers = nil
	this.regex = nil
	this.paramNames = nil
}

func (this *treeNode) match(path string) (Params, bool) {
	if this.regex != nil {
		return this.matchWithRegex(path)
	}
	if this.path == path {
		return nil, true
	}
	return nil, false
}

func (this *treeNode) matchWithRegex(path string) (Params, bool) {
	var mResult = this.regex.FindStringSubmatch(path)
	if len(mResult) == 0 {
		return nil, false
	}
	var mPath = mResult[0]
	if mPath != path {
		return nil, false
	}

	var param = make(Params)
	for index, item := range mResult {
		if index == 0 {
			continue
		}
		var name = this.paramNames[index-1]
		param.Set(name, item)
	}

	return param, true
}

func (this *treeNode) isPath() bool {
	return len(this.path) > 0 && len(this.handlers) > 0
}

func (this *treeNode) isValidPath(path string) bool {
	if len(this.handlers) > 0 && path == this.path {
		return true
	}
	return false
}

func (this *treeNode) isValidRegexPath() bool {
	if len(this.handlers) > 0 && this.regex != nil {
		return true
	}
	return false
}

func (this *treeNode) String() string {
	return fmt.Sprintf("{Depth:%d, Key:%s, Path:%s, isPath:%t}", this.depth, this.key, this.path, this.isPath())
}

func (this *treeNode) print() {
	for i := 0; i < this.depth; i++ {
		fmt.Print("-")
	}
	fmt.Println(this)
	for _, c := range this.subNodes {
		c.print()
	}
}
