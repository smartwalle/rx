package reverseproxy

import "testing"

func TestRewritePathPrefix(t *testing.T) {
	tests := []struct {
		name         string
		sourcePrefix string
		targetPrefix string
		path         string
		want         string
	}{
		{"前缀完全匹配", "/api", "/", "/api", "/"},
		{"请求路径带尾部斜杠", "/api", "/v1", "/api/", "/v1/"},
		{"匹配一级子路径", "/api", "/v1", "/api/users", "/v1/users"},
		{"匹配多级子路径", "/api", "/v1", "/api/users/1", "/v1/users/1"},

		{"相似前缀但不匹配", "/api", "/v1", "/api2/users", "/api2/users"},
		{"请求路径不包含源前缀", "/api", "/v1", "/users", "/users"},

		{"目标前缀为空且完全匹配", "/api", "", "/api", "/"},
		{"目标前缀为空且匹配子路径", "/api", "", "/api/users", "/users"},
		{"目标前缀为空且请求路径带尾部斜杠", "/api", "", "/api/", "/"},

		{"源前缀为空", "", "/v1", "/api", "/v1/api"},
		{"源前缀为空且请求根路径", "", "/v1", "/", "/v1"},
		{"源前缀为空且请求多级路径", "", "/v1", "/api/users", "/v1/api/users"},

		{"源前缀为根路径且请求子路径", "/", "/v1", "/api", "/v1/api"},
		{"源前缀为根路径且请求根路径", "/", "/v1", "/", "/v1"},
		{"源前缀为根路径且请求多级路径", "/", "/v1", "/api/users", "/v1/api/users"},

		{"目标前缀为根路径且完全匹配", "/api", "/", "/api", "/"},
		{"目标前缀为根路径且匹配子路径", "/api", "/", "/api/users", "/users"},
		{"目标前缀为根路径且请求路径带尾部斜杠", "/api", "/", "/api/", "/"},
		{"目标前缀为根路径且匹配子路径", "/api", "/", "/api/users", "/users"},
		{"目标前缀为根路径且匹配子路径带尾部斜杠", "/api", "/", "/api/users/", "/users/"},

		{"源前缀和目标前缀均为根路径", "/", "/", "/api", "/api"},
		{"源前缀和目标前缀均为根路径且请求根路径", "/", "/", "/", "/"},
		{"源前缀和目标前缀均为根路径且请求多级路径", "/", "/", "/api/users", "/api/users"},

		{"同时规范化源前缀和目标前缀", "api/", "v1/", "/api/users", "/v1/users"},
		{"规范化源前缀", "/api//v1/", "/v2", "/api/v1/users", "/v2/users"},
		{"规范化目标前缀", "/api", "/v1//users/", "/api/1", "/v1/users/1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rewriter := RewritePathPrefix(tt.sourcePrefix, tt.targetPrefix)

			if got := rewriter(tt.path); got != tt.want {
				t.Errorf(
					"RewritePathPrefix(%q, %q)(%q) = %q, want %q",
					tt.sourcePrefix,
					tt.targetPrefix,
					tt.path,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestRewritePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		in   string
		want string
	}{
		{"绝对目标路径", "/index.html", "/", "/index.html"},
		{"用户请求路径", "/index.html", "/users", "/index.html"},
		{"多级请求路径", "/index.html", "/api/users/1", "/index.html"},
		{"多级目标路径", "/api/v1/users", "/", "/api/v1/users"},
		{"多级目标路径和请求路径", "/api/v1/users", "/users", "/api/v1/users"},
		{"目标路径为根路径", "/", "/users", "/"},
		{"目标路径为空", "", "/users", "/"},
		{"规范化目标路径", "api//v1/", "/users", "/api/v1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rewriter := RewritePath(tt.path)

			if got := rewriter(tt.in); got != tt.want {
				t.Errorf(
					"RewritePath(%q)(%q) = %q, want %q",
					tt.path,
					tt.in,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestTrimPathPrefix(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		prefix  string
		want    string
		matched bool
	}{
		{"源前缀为空", "/api/users", "", "/api/users", true},
		{"源前缀为根路径", "/api/users", "/", "/api/users", true},
		{"请求根路径且源前缀为根路径", "/", "/", "/", true},
		{"请求路径与源前缀完全匹配", "/api", "/api", "/", true},
		{"请求路径匹配源前缀的子路径", "/api/users", "/api", "/users", true},
		{"请求路径与源前缀相似但不匹配", "/api2/users", "/api", "/api2/users", false},
		{"请求路径不包含源前缀", "/users", "/api", "/users", false},
		{"请求路径仅在源前缀后带斜杠", "/api/", "/api", "/", true},
		{"请求路径仅部分匹配源前缀", "/ap", "/api", "/ap", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, matched := trimPathPrefix(tt.path, tt.prefix)

			if got != tt.want {
				t.Errorf(
					"trimPathPrefix(%q, %q) = %q, want %q",
					tt.path,
					tt.prefix,
					got,
					tt.want,
				)
			}

			if matched != tt.matched {
				t.Errorf(
					"trimPathPrefix(%q, %q) matched = %t, want %t",
					tt.path,
					tt.prefix,
					matched,
					tt.matched,
				)
			}
		})
	}
}

func TestNormalizePathPrefix(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		want   string
	}{
		{"空路径", "", ""},
		{"根路径", "/", "/"},
		{"相对路径", "api", "/api"},
		{"绝对路径", "/api", "/api"},
		{"移除尾部斜杠", "/api/", "/api"},
		{"清理多余斜杠", "/api//v1/", "/api/v1"},
		{"清理当前目录", "./api", "/api"},
		{"清理父级目录", "/api/../v1", "/v1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizePathPrefix(tt.prefix); got != tt.want {
				t.Errorf(
					"normalizePathPrefix(%q) = %q, want %q",
					tt.prefix,
					got,
					tt.want,
				)
			}
		})
	}
}
