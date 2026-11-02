package rx

import (
	"log"
	"time"
)

// MidLogger 返回一个用于记录请求处理日志的中间件。
// 该中间件会在请求处理前记录开始时间，等待后续处理器执行完成后，
// 输出 HTTP 响应状态码、请求处理耗时、请求方法以及请求路径。
func MidLogger() HandlerFunc {
	return func(c *Context) {
		var begin = time.Now()

		c.Next()

		log.Printf("| %3d | %13v | %s - %s \n", c.Writer.Status(), time.Since(begin), c.Request.Method, c.Request.URL.Path)
	}
}

// MidErrorEncoder 返回一个用于设置当前请求错误响应编码器的中间件。
//
// Engine 支持配置全局默认错误响应编码器。通过该中间件，可以为不同的路由或路由组
// 设置独立的错误响应编码器，从而覆盖全局默认配置，实现更灵活的错误响应定制。
//
// 不同的 API 可能需要遵循不同的错误响应规范。例如，普通业务接口可以使用业务
// 自定义的错误响应格式，而 OAuth 接口则可以使用符合 OAuth 规范的错误响应格式。
// 通过在不同路由或路由组上注册该中间件，可以让不同类型的 API 使用各自的错误编码器。
//
// 如果 encoder 为 nil，则使用默认错误响应编码器。
//
// 配置的错误编码器仅作用于当前请求的 Context，不会修改 Rx 的全局默认配置。
func MidErrorEncoder(encoder ErrorEncoder) HandlerFunc {
	return func(c *Context) {
		if encoder == nil {
			encoder = defaultErrorEncoder
		}
		c.errorEncoder = encoder
	}
}
