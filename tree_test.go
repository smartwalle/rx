package rx

import (
	"fmt"
	"net/http"
	"testing"
)

func TestTree_Find(t *testing.T) {
	var tree = newMethodTree()
	tree.add("/", func(c *Context) {})
	tree.add("/t1", func(c *Context) {})
	tree.add("/t1/h1", func(c *Context) {})
	tree.add("/t1/h2", func(c *Context) {})
	tree.add("/t2/h1", func(c *Context) {})
	tree.add("/t2/h2", func(c *Context) {})
	tree.add("/t4", func(c *Context) {})

	tree.Print()

	var tests = []struct {
		m string
		p string
		c int
	}{
		{m: http.MethodGet, p: "/", c: 1},
		{m: http.MethodGet, p: "/t1", c: 1},
		{m: http.MethodGet, p: "/t1/h1", c: 1},
		{m: http.MethodGet, p: "/t1/h2", c: 1},
		{m: http.MethodGet, p: "/t2/h1", c: 1},
		{m: http.MethodGet, p: "/t2/h2", c: 1},
		{m: http.MethodGet, p: "/t4", c: 1},

		{m: http.MethodGet, p: "", c: 0},
		{m: http.MethodGet, p: "//", c: 0},
		{m: http.MethodGet, p: "/t1/", c: 0},
		{m: http.MethodGet, p: "/t2", c: 0},
		{m: http.MethodGet, p: "/t11", c: 0},
		{m: http.MethodGet, p: "/t3", c: 0},
		{m: http.MethodGet, p: "/t3/h1", c: 0},
		{m: http.MethodGet, p: "/t3/h1/", c: 0},
	}

	for _, test := range tests {
		if e := tree.find(test.p, false); len(e) != test.c {
			t.Errorf("%s - %s 的匹配结果应该为 %d, 实际为 %d", test.m, test.p, test.c, len(e))
		}
	}
}

func TestTree_Clean(t *testing.T) {
	var tree = newMethodTree()
	tree.add("/", func(c *Context) {})
	tree.add("/t1", func(c *Context) {})
	tree.add("/t1/h1", func(c *Context) {})
	tree.add("/t1/h2", func(c *Context) {})
	tree.add("/t2/h1", func(c *Context) {})
	tree.add("/t2/h2", func(c *Context) {})
	tree.add("/t4", func(c *Context) {})

	tree.Print()

	tree.clean("/t2/h1")
	tree.clean("/t2/h2")
	tree.clean("/t1")
	tree.clean("/t1/h1")
	tree.clean("/t1/h2")
	tree.clean("/t4")
	tree.clean("/")

	fmt.Println("-------------------")
	tree.Print()
}
