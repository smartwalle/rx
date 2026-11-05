package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/smartwalle/rx"
	"github.com/smartwalle/rx/reverseproxy"
	"github.com/smartwalle/rx/router"
	"github.com/smartwalle/rx/staticfile"
)

func main() {
	var r = router.New()

	// 示例导航页。
	r.GET("/", func(c *rx.Context) {
		c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, `<!DOCTYPE html>
<html lang="zh-CN">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>Rx Examples</title>
</head>
<body>
	<h1>Rx Examples</h1>

	<h2>普通路由</h2>
	<ul>
		<li><a href="/hello?name=Tom">Query 参数：/hello</a></li>
		<li><a href="/users/123">路径参数：/users/123</a></li>
	</ul>

	<h2>反向代理</h2>
	<ul>
		<li><a href="/jsonplaceholder/posts/1">JSONPlaceholder：获取文章</a></li>
		<li><a href="/anything/get">保留原路径</a></li>
		<li><a href="/proxy/replace/get">替换路径前缀</a></li>
		<li><a href="/proxy/strip/get">移除路径前缀</a></li>
		<li><a href="/proxy/add/get">添加路径前缀</a></li>
		<li><a href="/root-source/get">源前缀为根路径</a></li>
		<li><a href="/proxy/root/get">目标前缀为根路径</a></li>
		<li><a href="/proxy/fixed/get">重写为固定路径</a></li>
	</ul>

	<h2>错误处理</h2>
	<ul>
		<li><a href="/global-error">触发全局 ErrorEncoder</a></li>
		<li><a href="/oauth/error">触发路由组 ErrorEncoder</a></li>
	</ul>

	<h2>静态文件</h2>
	<ul>
		<li><a href="/static/main.go">查看 main.go</a></li>
	</ul>
</body>
</html>`)
	})

	// 普通路由：读取 Query 参数。
	r.GET("/hello", func(c *rx.Context) {
		var name = c.Request.URL.Query().Get("name")
		if name == "" {
			name = "world"
		}

		c.JSON(http.StatusOK, map[string]string{
			"message": "Hello " + name,
		})
	})

	r.POST("/h1", func(c *rx.Context) {
		fmt.Println("hh")
	})

	// 普通路由：读取路径参数。
	r.GET("/users/:id", func(c *rx.Context) {
		c.JSON(http.StatusOK, map[string]string{
			"id": c.Param("id"),
		})
	})

	// 反向代理：将请求转发到 JSONPlaceholder。
	// /jsonplaceholder/posts/1 -> /posts/1
	r.GET("/jsonplaceholder/*path", reverseproxy.NewHandler("https://jsonplaceholder.typicode.com", reverseproxy.RewritePathPrefix("/jsonplaceholder", "")))

	// 反向代理：不重写请求路径。
	// /anything/get -> https://httpbin.org/anything/get
	r.GET("/anything/*path", reverseproxy.NewHandler("https://httpbin.org", nil))

	// 反向代理：替换路径前缀。
	// /proxy/replace/get -> /anything/get
	r.GET("/proxy/replace/*path", reverseproxy.NewHandler("https://httpbin.org", reverseproxy.RewritePathPrefix("/proxy/replace", "/anything")))

	// 反向代理：移除路径前缀。
	// /proxy/strip/get -> /get
	r.GET("/proxy/strip/*path", reverseproxy.NewHandler("https://httpbin.org", reverseproxy.RewritePathPrefix("/proxy/strip", "")))

	// 反向代理：添加路径前缀。
	// /proxy/add/get -> /anything/proxy/add/get
	r.GET("/proxy/add/*path", reverseproxy.NewHandler("https://httpbin.org", reverseproxy.RewritePathPrefix("", "/anything")))

	// 反向代理：源前缀为根路径。
	// /root-source/get -> /anything/root-source/get
	r.GET("/root-source/*path", reverseproxy.NewHandler("https://httpbin.org", reverseproxy.RewritePathPrefix("/", "/anything")))

	// 反向代理：目标前缀为根路径。
	// /proxy/root/get -> /get
	r.GET("/proxy/root/*path", reverseproxy.NewHandler("https://httpbin.org", reverseproxy.RewritePathPrefix("/proxy/root", "/")))

	// 反向代理：将所有请求路径重写为固定路径。
	// 无论访问什么子路径，实际请求路径始终为 /anything/get。
	r.GET("/proxy/fixed/*path", reverseproxy.NewHandler("https://httpbin.org", reverseproxy.RewritePath("/anything/get")))

	// 触发全局错误编码器。
	r.GET("/global-error", func(c *rx.Context) {
		c.AbortWithError(errors.New("this error uses the global encoder"))
	})

	// 为 OAuth 路由组设置独立的错误编码器。
	var oauth = r.Group("/oauth", rx.MidErrorEncoder(func(c *rx.Context, err error) {
		c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_request",
			"error_description": err.Error(),
		})
	}))

	oauth.GET("/error", func(c *rx.Context) {
		c.AbortWithError(errors.New("the request is missing a required parameter"))
	})

	// 静态文件：访问当前目录下的 main.go。
	r.GET("/static/*filepath", staticfile.NewHandler("./"))

	var s = rx.New()
	s.Use(rx.MidLogger())
	s.UseRouter(r)

	// 全局错误编码器。
	s.UseErrorEncoder(func(c *rx.Context, err error) {
		c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	})

	s.UseNotFoundHandler(func(c *rx.Context) {
		c.String(-1, "访问路由不存在")
	})

	s.UseMethodNotAllowedHandler(func(c *rx.Context) {
		c.String(-1, "请求方法不支持")
	})

	if err := http.ListenAndServe(":9900", s); err != nil {
		log.Fatalln(err)
	}
}
