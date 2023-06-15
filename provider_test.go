package rx_test

import (
	"github.com/smartwalle/rx"
	"net/http/httptest"
	"testing"
)

func TestListProvider_Match(t *testing.T) {
	var provider = rx.NewListProvider()
	// 要想能够正常匹配 /user, 需要在添加 /u 之前添加 /user
	provider.Add("/user", []string{"http://127.0.0.1:9920", "http://127.0.0.1:9921"})
	provider.Add("/u", []string{"http://127.0.0.1:9910", "http://127.0.0.1:9911"})

	var tests = []struct {
		path    string
		pattern string
	}{
		{
			path:    "/u",
			pattern: "/u",
		},
		{
			path:    "/u/u",
			pattern: "/u",
		},
		{
			path:    "/us",
			pattern: "/u",
		},
		{
			path:    "/us/u",
			pattern: "/u",
		},
		{
			path:    "/use",
			pattern: "/u",
		},
		{
			path:    "/use/u",
			pattern: "/u",
		},
		{
			path:    "/user",
			pattern: "/user",
		},
		{
			path:    "/user/u",
			pattern: "/user",
		},
	}

	for _, test := range tests {
		var req = httptest.NewRequest("GET", "http://127.0.0.1"+test.path, nil)
		var route, _ = provider.Match(req)
		if route.Pattern() != test.pattern {
			t.Fatalf("期望: %s，实际：%s \n", test.pattern, route.Pattern())
		}
	}

}
