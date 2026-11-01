package reverseproxy

import (
	"net/http/httputil"
	"net/url"
	stdpath "path"
	"strings"

	"github.com/smartwalle/rx"
)

// PathRewriter 用于重写请求路径。
type PathRewriter func(path string) string

// RewritePathPrefix 返回一个路径重写函数，将请求路径中匹配的源前缀替换为目标前缀。
// sourcePrefix 和 targetPrefix 均会进行路径规范化。
// 仅当请求路径与 sourcePrefix 完全匹配，或前缀后紧跟 "/" 时才执行替换。
//
// 示例：
//
//	rewriter := RewritePathPrefix("/api", "/v1")
//	rewriter("/api")              // "/v1"           前缀完全匹配
//	rewriter("/api/")             // "/v1/"          前缀后带尾部斜杠
//	rewriter("/api/users")        // "/v1/users"     子路径匹配
//	rewriter("/api/users/1")      // "/v1/users/1"   多级子路径匹配
//	rewriter("/api2/users")       // "/api2/users"   相似前缀，不匹配
//	rewriter("/users")            // "/users"        不包含源前缀
//
//	rewriter = RewritePathPrefix("/api", "")
//	rewriter("/api")              // "/"             目标前缀为空
//	rewriter("/api/users")        // "/users"        移除源前缀
//
//	rewriter = RewritePathPrefix("", "/v1")
//	rewriter("/api")              // "/v1/api"       源前缀为空
//	rewriter("/")                 // "/v1"           请求根路径
//
//	rewriter = RewritePathPrefix("/", "/v1")
//	rewriter("/api")              // "/v1/api"       源前缀为根路径
//	rewriter("/")                 // "/v1"           请求根路径
//
//	rewriter = RewritePathPrefix("/api", "/")
//	rewriter("/api")              // "/"             目标前缀为根路径
//	rewriter("/api/users")        // "/users"        子路径匹配
//	rewriter("/api/users/")       // "/users/"       子路径带尾部斜杠
func RewritePathPrefix(sourcePrefix, targetPrefix string) PathRewriter {
	// 规范化源前缀和目标前缀。
	sourcePrefix = normalizePathPrefix(sourcePrefix)
	targetPrefix = normalizePathPrefix(targetPrefix)

	return func(path string) string {
		// 移除匹配的源前缀，未匹配时保留原路径。
		matchedPath, matched := trimPathPrefix(path, sourcePrefix)
		if !matched {
			return path
		}

		// 目标前缀为空时，直接返回移除源前缀后的路径。
		if targetPrefix == "" {
			return matchedPath
		}

		// 移除源前缀后路径为根路径时，直接使用目标前缀。
		// 源前缀为空时，请求根路径也适用此规则。
		if matchedPath == "/" && (sourcePrefix == "" || path == sourcePrefix) {
			return targetPrefix
		}

		// 目标前缀为根路径时，直接使用移除源前缀后的路径。
		if targetPrefix == "/" {
			return matchedPath
		}

		// 将目标前缀与剩余路径拼接。
		return targetPrefix + matchedPath
	}
}

// RewritePath 返回一个路径重写函数，将所有请求路径统一重写为指定路径。
// 指定路径会进行路径规范化；为空时，默认重写为根路径 "/"。
//
// 示例：
//
//	rewriter := RewritePath("/index.html")
//	rewriter("/")                  // "/index.html"
//	rewriter("/users")             // "/index.html"
//	rewriter("/api/users/1")       // "/index.html"
//
//	rewriter = RewritePath("/api/v1/users")
//	rewriter("/")                  // "/api/v1/users"
//	rewriter("/users")             // "/api/v1/users"
//
//	rewriter = RewritePath("/")
//	rewriter("/users")             // "/"
//
//	rewriter = RewritePath("")
//	rewriter("/users")             // "/"
func RewritePath(path string) PathRewriter {
	path = normalizePathPrefix(path)
	if path == "" {
		path = "/"
	}
	return func(string) string {
		return path
	}
}

// trimPathPrefix 从请求路径中移除指定的路径前缀。
func trimPathPrefix(path, prefix string) (string, bool) {
	// 空前缀或根路径前缀不需要移除。
	if prefix == "" || prefix == "/" {
		return path, true
	}

	// 请求路径与前缀完全匹配时，返回根路径。
	if path == prefix {
		return "/", true
	}

	// 仅匹配完整的路径段，避免将相似前缀误判为匹配。
	if strings.HasPrefix(path, prefix+"/") {
		return strings.TrimPrefix(path, prefix), true
	}

	// 前缀不匹配时，保留原路径。
	return path, false
}

// normalizePathPrefix 将路径前缀规范化为以 "/" 开头的路径。
// 空字符串保持不变，非空路径通过 path.Join 清理多余的路径分隔符及路径元素。
func normalizePathPrefix(prefix string) string {
	if prefix == "" {
		return ""
	}
	return stdpath.Join("/", prefix)
}

// NewHandler 创建一个反向代理处理器，将请求转发到指定目标地址。
// rewriter 用于重写转发请求的路径，可以为 nil，表示保留原请求路径。
// target 无法解析时会触发 panic。
func NewHandler(target string, rewriter PathRewriter) rx.HandlerFunc {
	var targetURL, err = url.Parse(target)
	if err != nil {
		panic(err)
	}

	var proxy = &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(targetURL)
			r.Out.Host = targetURL.Host

			if rewriter != nil {
				r.Out.URL.Path = rewriter(r.Out.URL.Path)
				r.Out.URL.RawPath = ""
			}
		},
	}

	return func(c *rx.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
