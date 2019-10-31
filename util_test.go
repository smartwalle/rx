package rx

import (
	"fmt"
	"testing"
)

func Test_CleanPath(t *testing.T) {
	var tests = []struct {
		src string
		dst string
	}{
		{src: "", dst: "/"},
		{src: "/", dst: "/"},
		{src: "/p1", dst: "/p1"},
		{src: "/p1/", dst: "/p1"},
		{src: "p1", dst: "/p1"},
		{src: "p1/", dst: "/p1"},
		{src: "p1/p2", dst: "/p1/p2"},
		{src: "/p1/p2", dst: "/p1/p2"},
	}

	for _, test := range tests {
		if r := CleanPath(test.src); r != test.dst {
			t.Errorf("%s 转换之后应该得到 %s, 实际结果为 %s \n", test.src, test.dst, r)
		} else {
			fmt.Println(r)
		}
	}
}
