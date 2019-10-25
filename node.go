package rx

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	defaultWildcard = `([\w]+)`
)

type pathNode struct {
	key      string
	path     string
	isPath   bool
	depth    int
	children map[string]*pathNode
	handlers []HandlerFunc

	regex      *regexp.Regexp
	paramNames []string
}

func newPathNode(key string, depth int) *pathNode {
	var n = &pathNode{}
	n.key = key
	n.depth = depth
	n.children = make(map[string]*pathNode)
	return n
}

func (this *pathNode) prepare(path string, handlers ...HandlerFunc) {
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
			pattern = pattern + "/" + defaultWildcard
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
