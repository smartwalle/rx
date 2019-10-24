package rx

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	defaultWild = `([\w]+)`
)

type Node struct {
	key      string
	path     string
	isPath   bool
	depth    int
	children map[string]*Node
	handlers []HandlerFunc

	regex      *regexp.Regexp
	paramNames []string
}

func newNode(key string, depth int) *Node {
	var n = &Node{}
	n.key = key
	n.depth = depth
	n.children = make(map[string]*Node)
	return n
}

func (this *Node) prepare(path string, handlers ...HandlerFunc) {
	this.isPath = true
	this.path = path
	this.handlers = handlers

	var paths = splitPath(path)
	var pattern = ""
	var paramsNames = make([]string, 0, len(paths))

	for _, p := range paths {
		if p == "" {
			continue
		}

		var strLen = len(p)
		var lastChar = lastChar(p)
		var firstChar = firstChar(p)

		if firstChar == ':' {
			var name = p[1:strLen]
			pattern = pattern + "/" + defaultWild
			paramsNames = append(paramsNames, name)
		} else if firstChar == '{' && lastChar == '}' {
			var subStrList = strings.Split(p[1:strLen-1], ":")
			paramsNames = append(paramsNames, subStrList[0])
			pattern = pattern + "/" + subStrList[1]
		} else {
			pattern = pattern + "/" + p
		}
	}

	this.regex = regexp.MustCompile(pattern)
	this.paramNames = paramsNames
}

func (this *Node) match(path string) (Params, bool) {
	if this.regex != nil {
		return this.matchWithRegex(path)
	}
	if this.path == path {
		return nil, true
	}
	return nil, false
}

func (this *Node) matchWithRegex(path string) (Params, bool) {
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

func (this *Node) String() string {
	return fmt.Sprintf("{Key:%s Path:%s}", this.key, this.path)
}

func (this *Node) Print() {
	for i := 0; i < this.depth; i++ {
		fmt.Print("-")
	}
	fmt.Println(this.String())
	for _, c := range this.children {
		c.Print()
	}
}
