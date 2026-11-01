package reverseproxy

import (
	"net/http/httputil"
	"net/url"
	stdpath "path"
	"strings"

	"github.com/smartwalle/rx"
)

type PathRewriter func(path string) string

// RewritePathPrefix 替换路径前缀。
func RewritePathPrefix(sourcePrefix, targetPrefix string) PathRewriter {
	sourcePrefix = normalizePathPrefix(sourcePrefix)
	targetPrefix = normalizePathPrefix(targetPrefix)

	return func(path string) string {
		path = trimPathPrefix(path, sourcePrefix)
		if targetPrefix == "" {
			return path
		}
		if path == "/" {
			return targetPrefix
		}
		return targetPrefix + path
	}
}

// RewritePath 重写整个请求路径。
func RewritePath(path string) PathRewriter {
	path = normalizePathPrefix(path)
	if path == "" {
		path = "/"
	}
	return func(string) string {
		return path
	}
}

func trimPathPrefix(path, prefix string) string {
	if prefix == "" || prefix == "/" {
		return path
	}

	if path == prefix {
		return "/"
	}

	if strings.HasPrefix(path, prefix+"/") {
		return strings.TrimPrefix(path, prefix)
	}

	return path
}

func normalizePathPrefix(prefix string) string {
	if prefix == "" {
		return ""
	}
	return stdpath.Join("/", prefix)
}

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
