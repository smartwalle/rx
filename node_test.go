package rx

import (
	"testing"
)

func TestNode_Match(t *testing.T) {
	var handlers = HandlerChain{}
	handlers = append(handlers, func(c *Context) {})

	var tests = []struct {
		path        string
		testPath    string
		match       bool
		numOfParams int
	}{
		// 普通匹配
		{path: "/", testPath: "/", match: true, numOfParams: 0},
		{path: "/r1", testPath: "/r1", match: true, numOfParams: 0},
		{path: "/r1", testPath: "/r11", match: false, numOfParams: 0},

		{path: "/r1", testPath: "/r1/", match: false, numOfParams: 0},
		{path: "/r1/", testPath: "/r1/", match: true, numOfParams: 0},
		{path: "/r1", testPath: CleanPath("/r1/"), match: true, numOfParams: 0},

		{path: "/r1/r2", testPath: "/r1/r2", match: true, numOfParams: 0},
		{path: "/r1/r2", testPath: "/r1/r2/r3", match: false, numOfParams: 0},
		{path: "/r1/r2/r3", testPath: "/r1/r2", match: false, numOfParams: 0},

		// 正则匹配
		{path: "/:id", testPath: "/1", match: true, numOfParams: 1},
		{path: "/r1/:id", testPath: "/r1/2", match: true, numOfParams: 1},
		{path: "/r1/:id", testPath: "/r1/myid", match: true, numOfParams: 1},
		{path: "/r1/:name", testPath: "/r1/唯一标识", match: true, numOfParams: 1},

		{path: "/r1/:id/r2", testPath: "/r1/2/r2", match: true, numOfParams: 1},
		{path: "/r1/:id/*action", testPath: "/r1/2/r2", match: true, numOfParams: 2},
		{path: "/r1/:id/*action", testPath: "/r1/2/r2/r3", match: true, numOfParams: 2},
		{path: "/r1/:id/*action", testPath: "/r1/2/r2/r3/r4", match: true, numOfParams: 2},

		{path: "/r1/{id:([\\d]+)}", testPath: "/r1/2", match: true, numOfParams: 1},
		{path: "/r1/{id:([\\d]+)}", testPath: "/r1/2s", match: false, numOfParams: 0},
		{path: "/r1/{id:([\\d]+)}", testPath: "/r1/ss", match: false, numOfParams: 0},
	}

	for _, test := range tests {
		var n = newPathNode("test", 1, 1)
		n.prepare(test.path, handlers)

		var params, ok = n.match(test.testPath, nil)
		if ok != test.match || len(params) != test.numOfParams {
			t.Errorf("%s - %s 的匹配结果应该为 %t - %d, 实际为 %t - %d", test.path, test.testPath, test.match, test.numOfParams, ok, len(params))
		}
	}
}
