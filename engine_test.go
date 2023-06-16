package rx_test

import (
	"github.com/smartwalle/rx"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Get(t *testing.T, server *httptest.Server, path string) *http.Response {
	rsp, err := server.Client().Get(server.URL + path)
	if err != nil {
		t.Fatal(err)
	}
	return rsp
}

func NewBackend() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/200":
			w.WriteHeader(http.StatusOK)
		case "/2001":
			w.WriteHeader(http.StatusOK)
		case "/201":
			w.WriteHeader(http.StatusCreated)
		case "/2011":
			w.WriteHeader(http.StatusCreated)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
}

func NewProvider(host string) rx.RouteProvider {
	var provider = rx.NewListProvider()
	provider.Add("/200", []string{host})
	provider.Add("/201", []string{host})
	provider.Add("/400", []string{host})
	return provider
}

func TestEngine_D1(t *testing.T) {
	var backend = NewBackend()
	defer backend.Close()

	var engine = rx.New()
	engine.Load(NewProvider(backend.URL))

	frontend := httptest.NewServer(engine)
	defer frontend.Close()

	var tests = []struct {
		path   string
		expect int
	}{
		{
			path:   "/200", // 有注册该路由，目标服务器返回的是 http.StatusOK
			expect: http.StatusOK,
		},
		{
			path:   "/2001", // 匹配到 /200
			expect: http.StatusOK,
		},
		{
			path:   "/201", // 有注册该路由，目标服务器返回的是 http.StatusCreated
			expect: http.StatusCreated,
		},
		{
			path:   "/2011", // 匹配到 /201
			expect: http.StatusCreated,
		},
		{
			path:   "/400", // 有注册该路由，目标服务器返回的是 http.StatusBadRequest
			expect: http.StatusBadRequest,
		},
		{
			path:   "/4001", // 匹配到 /300
			expect: http.StatusBadRequest,
		},
		{
			path:   "/502", // 未注册该路由，所以返回 http.StatusBadGateway
			expect: http.StatusBadGateway,
		},
	}

	for _, test := range tests {
		var rsp = Get(t, frontend, test.path)
		if rsp.StatusCode != test.expect {
			t.Fatalf("访问：%s 期望: %d，实际：%d \n", test.path, test.expect, rsp.StatusCode)
		}
	}
}

func TestEngine_Abort(t *testing.T) {
	var backend = NewBackend()
	defer backend.Close()

	var engine = rx.New()
	engine.Load(NewProvider(backend.URL))

	engine.Use(func(c *rx.Context) {
		switch c.Request.URL.Path {
		case "/200":
			c.AbortWithStatus(http.StatusUnauthorized)
			c.AbortWithStatus(http.StatusOK)
			c.AbortWithStatus(http.StatusCreated)
		case "/2001":
			c.Status(http.StatusMultiStatus)
		case "/2002":
			c.Status(http.StatusAccepted)
			c.Abort()
		case "/2003":
			c.Abort()
			c.Status(http.StatusAlreadyReported)
		case "/2004":
			c.Status(http.StatusConflict)
		case "/201":
			c.Next()
			c.AbortWithStatus(http.StatusBadRequest)
		case "/400":
			c.AbortWithStatus(http.StatusCreated)
			c.Abort()
		}
	}, func(c *rx.Context) {
		switch c.Request.URL.Path {
		case "/200":
			t.Fatal("不应该执行到这里")
		case "/2002":
			t.Fatal("不应该执行到这里")
		case "/2003":
			t.Fatal("不应该执行到这里")
		case "/2004":
			c.Abort()
		case "/400":
			t.Fatal("不应该执行到这里")
		}
	})

	frontend := httptest.NewServer(engine)
	defer frontend.Close()

	var tests = []struct {
		path   string
		expect int
	}{
		{
			path:   "/200", // 有注册该路由，但是在 middleware 中将返回值调整为 http.StatusUnauthorized
			expect: http.StatusUnauthorized,
		},
		{
			path:   "/2001", // 匹配到 /200，在 middleware 中虽然有将返回值调整为 http.StatusMultiStatus，但是没有 abort，仍然以目标服务器的返回值为准
			expect: http.StatusOK,
		},
		{
			path:   "/2002", // 匹配到 /200，在 middleware 中将返回值调整为 http.StatusAccepted 并且 abort
			expect: http.StatusAccepted,
		},
		{
			path:   "/2003", // 匹配到 /200，在 middleware 中将返回值调整为 http.StatusAlreadyReported 并且 abort
			expect: http.StatusAlreadyReported,
		},
		{
			path:   "/2004", // 匹配到 /200，在 middleware 中将返回值调整为 http.StatusConflict 并且 abort
			expect: http.StatusConflict,
		},
		{
			path:   "/201", // 匹配到 /201，目标服务器返回值为 http.StatusCreated，但是在 middleware 中将返回值调整为 http.StatusBadRequest
			expect: http.StatusBadRequest,
		},
		{
			path:   "/400", // 有注册该路由，但是在 middleware 中将返回值调整为 http.StatusCreated
			expect: http.StatusCreated,
		},
	}

	for _, test := range tests {
		var rsp = Get(t, frontend, test.path)
		if rsp.StatusCode != test.expect {
			t.Fatalf("访问：%s 期望: %d，实际：%d \n", test.path, test.expect, rsp.StatusCode)
		}
	}
}

func TestEngine_NoRouteAndAbort(t *testing.T) {
	var backend = NewBackend()
	defer backend.Close()

	var engine = rx.New()
	engine.Load(NewProvider(backend.URL))

	engine.Use(func(c *rx.Context) {
		switch c.Request.URL.Path {
		case "/5022":
			c.AbortWithStatus(http.StatusUnauthorized)
		case "/5023":
			c.Abort()
		}
	}, func(c *rx.Context) {
		switch c.Request.URL.Path {
		case "/5022":
			t.Fatal("不应该执行到这里")
		case "/5023":
			t.Fatal("不应该执行到这里")
		}
	})

	// 没有注册的路由默认返回状态码为 http.StatusBadGateway
	engine.NoRoute(func(c *rx.Context) {
	}, func(c *rx.Context) {
		switch c.Request.URL.Path {
		case "/5021":
			// 在此调整为返回 http.StatusInternalServerError
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	}, func(c *rx.Context) {
		switch c.Request.URL.Path {
		case "/5021":
			t.Fatal("不应该执行到这里")
		case "/5022":
			t.Fatal("不应该执行到这里")
		case "/5023":
			t.Fatal("不应该执行到这里")
		}
	})

	frontend := httptest.NewServer(engine)
	defer frontend.Close()

	var tests = []struct {
		path   string
		expect int
	}{
		{
			path:   "/200", // 有注册该路由，目标服务器返回的是 http.StatusOK
			expect: http.StatusOK,
		},
		{
			path:   "/400", // 有注册该路由，目标服务器返回的是 http.StatusBadRequest
			expect: http.StatusBadRequest,
		},
		{
			path:   "/502", // 未注册该路由，默认返回 http.StatusBadGateway
			expect: http.StatusBadGateway,
		},
		{
			path:   "/5021", // 未注册该路由，默认返回 http.StatusBadGateway，但是在 NoRoute 中将返回值调整为 http.StatusInternalServerError
			expect: http.StatusInternalServerError,
		},
		{
			path:   "/5022", // 未注册该路由，默认返回 http.StatusBadGateway，但是在 middleware 中将返回值调整为 http.StatusUnauthorized
			expect: http.StatusUnauthorized,
		},
		{
			path:   "/5023", // 未注册该路由，默认返回 http.StatusBadGateway，在 middleware 中中断时未指定状态码，所以返回 http.StatusBadGateway
			expect: http.StatusBadGateway,
		},
	}

	for _, test := range tests {
		var rsp = Get(t, frontend, test.path)
		if rsp.StatusCode != test.expect {
			t.Fatalf("访问：%s 期望: %d，实际：%d \n", test.path, test.expect, rsp.StatusCode)
		}
	}
}

func TestEngine_Error(t *testing.T) {
	var engine = rx.New()
	engine.Load(NewProvider("http://127.0.0.1:1100"))

	engine.Use(func(c *rx.Context) {
		switch c.Request.URL.Path {
		case "/2002":
			c.AbortWithStatus(http.StatusUnauthorized)
		case "/2003":
			c.Abort()
		}
	})

	engine.ErrorHandler(func(c *rx.Context, err error) {
		switch c.Request.URL.Path {
		case "/200":
			c.AbortWithStatus(http.StatusNotImplemented)
		case "/400":
			c.AbortWithStatus(http.StatusServiceUnavailable)
		}
	})

	frontend := httptest.NewServer(engine)
	defer frontend.Close()

	var tests = []struct {
		path   string
		expect int
	}{
		{
			path:   "/200", // 有注册该路由，目标服务器无法访问触发错误，在 ErrorHandler 中将返回值调整为 http.StatusNotImplemented
			expect: http.StatusNotImplemented,
		},
		{
			path:   "/2001", // 匹配到 /200，目标服务器无法访问触发错误，ErrorHandler 中没有处理该路由，所以返回默认 http.StatusInternalServerError
			expect: http.StatusInternalServerError,
		},
		{
			path:   "/2002", // 匹配到 /200，在 middleware 中将返回值调整为 http.StatusUnauthorized
			expect: http.StatusUnauthorized,
		},
		{
			path:   "/2003", // 匹配到 /200，在 middleware 中中断时未指定状态码，所以返回 http.StatusOK
			expect: http.StatusOK,
		},
		{
			path:   "/400", // 有注册该路由，目标服务器无法访问触发错误，在 ErrorHandler 中将返回值调整为 http.StatusServiceUnavailable
			expect: http.StatusServiceUnavailable,
		},
		{
			path:   "/502", // 未注册该路由，默认返回 http.StatusBadGateway
			expect: http.StatusBadGateway,
		},
	}

	for _, test := range tests {
		var rsp = Get(t, frontend, test.path)
		if rsp.StatusCode != test.expect {
			t.Fatalf("访问：%s 期望: %d，实际：%d \n", test.path, test.expect, rsp.StatusCode)
		}
	}
}

func NewUserBackend() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user/list":
			w.WriteHeader(http.StatusOK)
		case "/user/123":
			w.WriteHeader(http.StatusOK)
		case "/user/456":
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
}

func NewOrderBackend() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/order/list":
			w.WriteHeader(http.StatusOK)
		case "/order/123":
			w.WriteHeader(http.StatusOK)
		case "/order/456":
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
}

func TestEngine_D2(t *testing.T) {
	var userBackend = NewUserBackend()
	defer userBackend.Close()

	var orderBackend = NewOrderBackend()
	defer orderBackend.Close()

	var provider = rx.NewListProvider()
	provider.Add("/user", []string{userBackend.URL})
	provider.Add("/order", []string{orderBackend.URL})

	var engine = rx.New()
	engine.Load(provider)

	engine.Use(func(c *rx.Context) {
		c.Set("kyo", "first")
	})
	engine.Use(func(c *rx.Context) {
		var _, ok = c.Get("kyo")
		if !ok {
			t.Fatal("应该存在 Key：kyo")
		}
	})

	frontend := httptest.NewServer(engine)
	defer frontend.Close()

	var tests = []struct {
		path   string
		expect int
	}{
		{
			path:   "/user/list", // 匹配到 /user，目标服务器存在 /user/list 且返回的是 http.StatusOK
			expect: http.StatusOK,
		},
		{
			path:   "/user/lists", // 匹配到 /user，目标服务器不存在 /user/lists 且返回的是 http.StatusBadRequest
			expect: http.StatusBadRequest,
		},
		{
			path:   "/user/123", // 匹配到 /user，目标服务器存在 /user/123 且返回的是 http.StatusOK
			expect: http.StatusOK,
		},
		{
			path:   "/user/456", // 匹配到 /user，目标服务器存在 /user/456 且返回的是 http.StatusOK
			expect: http.StatusOK,
		},
		{
			path:   "/user/789", // 匹配到 /user，目标服务器不存在 /user/789 且返回的是 http.StatusBadRequest
			expect: http.StatusBadRequest,
		},
		{
			path:   "/order/list",
			expect: http.StatusOK,
		},
		{
			path:   "/order/lists",
			expect: http.StatusBadRequest,
		},
		{
			path:   "/order/123",
			expect: http.StatusOK,
		},
		{
			path:   "/order/456",
			expect: http.StatusOK,
		},
		{
			path:   "/order/789",
			expect: http.StatusBadRequest,
		},
		{
			path:   "/book/list", // 未注册该路由，默认返回 http.StatusBadGateway
			expect: http.StatusBadGateway,
		},
	}

	for _, test := range tests {
		var rsp = Get(t, frontend, test.path)
		if rsp.StatusCode != test.expect {
			t.Fatalf("访问：%s 期望: %d，实际：%d \n", test.path, test.expect, rsp.StatusCode)
		}
	}
}

func TestEngine_FixPath(t *testing.T) {
	var userBackend = NewUserBackend()
	defer userBackend.Close()

	var provider = rx.NewListProvider()
	provider.Add("/api/user", []string{userBackend.URL})

	var engine = rx.New()
	engine.Load(provider)

	engine.Use(func(c *rx.Context) {
		switch c.Request.URL.Path {
		case "/api/user/456":
		default:
			// 去除 URL Path 中的 /api
			c.Request.URL.Path = strings.TrimPrefix(c.Request.URL.Path, "/api")
		}
	})

	frontend := httptest.NewServer(engine)
	defer frontend.Close()

	var tests = []struct {
		path   string
		expect int
	}{
		{
			path:   "/api/user/list", // 匹配到 /api/user，在 middleware 中将 URL Path 中的 /api 去掉，目标服务器存在 /user/list 且返回的是 http.StatusOK
			expect: http.StatusOK,
		},
		{
			path:   "/api/user/lists", // 匹配到 /api/user，在 middleware 中将 URL Path 中的 /api 去掉，目标服务器不存在 /user/lists 且返回的是 http.StatusBadRequest
			expect: http.StatusBadRequest,
		},
		{
			path:   "/api/user/123", // 匹配到 /api/user，在 middleware 中将 URL Path 中的 /api 去掉，目标服务器存在 /user/123 且返回的是 http.StatusOK
			expect: http.StatusOK,
		},
		{
			path:   "/api/user/456", // 匹配到 /api/user，在 middleware 中没有将 URL Path 中的 /api 去掉，目标服务器不存在 /api/user/456 且返回的是 http.StatusBadRequest
			expect: http.StatusBadRequest,
		},
		{
			path:   "/api/user/789", // 匹配到 /api/user，在 middleware 中将 URL Path 中的 /api 去掉，目标服务器不存在 /user/789 且返回的是 http.StatusBadRequest
			expect: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		var rsp = Get(t, frontend, test.path)
		if rsp.StatusCode != test.expect {
			t.Fatalf("访问：%s 期望: %d，实际：%d \n", test.path, test.expect, rsp.StatusCode)
		}
	}
}
