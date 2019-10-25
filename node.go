package rx

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	wildcard1 = `([^\s/]+)`
	wildcard2 = `([\S]+)`
)

type pathNodes map[string]*pathNode

func (this pathNodes) add(node *pathNode) {
	if node == nil {
		return
	}
	this[node.key] = node
}

func (this pathNodes) get(key string) *pathNode {
	return this[key]
}

func (this pathNodes) del(key string) {
	delete(this, key)
}

type pathNode struct {
	key      string
	depth    int
	children pathNodes

	path     string
	isPath   bool
	handlers HandlerChain

	regex      *regexp.Regexp
	paramNames []string
}

func newPathNode(key string, depth int) *pathNode {
	var n = &pathNode{}
	n.key = key
	n.depth = depth
	n.children = make(pathNodes)
	return n
}

func (this *pathNode) reset() {
	this.path = ""
	this.isPath = false
	this.handlers = nil
	this.regex = nil
	this.paramNames = nil
}

func (this *pathNode) add(node *pathNode) {
	this.children.add(node)
}

func (this *pathNode) get(key string) *pathNode {
	return this.children.get(key)
}

func (this *pathNode) remove(key string) {
	this.children.del(key)
}

func (this *pathNode) prepare(path string, handlers HandlerChain) {
	this.path = path
	this.isPath = true
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

func (this *pathNode) match(path string) (Params, bool) {
	if this.regex != nil {
		return this.matchWithRegex(path)
	}
	if this.path == path {
		return nil, true
	}
	return nil, false
}

func (this *pathNode) matchWithRegex(path string) (Params, bool) {
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

func (this *pathNode) String() string {
	return fmt.Sprintf("{Key:%s Path:%s}", this.key, this.path)
}

func (this *pathNode) Print() {
	for i := 0; i < this.depth; i++ {
		fmt.Print("-")
	}
	fmt.Println(this.String())
	for _, c := range this.children {
		c.Print()
	}
}
