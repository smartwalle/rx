package rx_test

import (
	"fmt"
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
		case BuildPath(http.StatusOK):
			w.WriteHeader(http.StatusOK)
		case BuildPath(2001):
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
}

func BuildPath(code int) string {
	return fmt.Sprintf("/%d", code)
}

func NewProvider(host string) rx.RouteProvider {
	var provider = rx.NewListProvider()
	provider.Add(BuildPath(http.StatusOK), []string{host})
	provider.Add(BuildPath(http.StatusBadRequest), []string{host})
	return provider
}

func TestEngine_T1(t *testing.T) {
	var backend = NewBackend()
	defer backend.Close()

	var engine = rx.New()
	engine.Load(NewProvider(backend.URL))

	frontend := httptest.NewServer(engine)
	defer frontend.Close()

	var tests = []struct {
		path   int
		expect int
	}{
		{
			path:   http.StatusOK,
			expect: http.StatusOK,
		},
		{
			path:   2001,
			expect: http.StatusOK,
		},
		{
			path:   http.StatusBadRequest,
			expect: http.StatusBadRequest,
		},
		{
			path:   4001,
			expect: http.StatusBadRequest,
		},
		{
			path:   http.StatusBadGateway,
			expect: http.StatusBadGateway,
		},
	}

	for _, test := range tests {
		var rsp = Get(t, frontend, BuildPath(test.path))
		if rsp.StatusCode != test.expect {
			t.Fatalf("访问：%s 期望: %d，实际：%d \n", BuildPath(test.path), test.expect, rsp.StatusCode)
		}
	}
}

func TestEngine_NoRoute(t *testing.T) {
	var backend = NewBackend()
	defer backend.Close()

	var engine = rx.New()
	engine.Load(NewProvider(backend.URL))

	// 没有注册的路由默认返回状态码为 http.StatusBadGateway，在此调整为返回 http.StatusInternalServerError
	engine.NoRoute(func(c *rx.Context) {
	}, func(c *rx.Context) {
		c.AbortWithStatus(http.StatusInternalServerError)
	}, func(c *rx.Context) {
		t.Fatal("不应该执行到这里")
	})

	frontend := httptest.NewServer(engine)
	defer frontend.Close()

	var tests = []struct {
		path   int
		expect int
	}{
		{
			path:   http.StatusOK,
			expect: http.StatusOK,
		},
		{
			path:   http.StatusBadRequest,
			expect: http.StatusBadRequest,
		},
		{
			path:   http.StatusBadGateway,
			expect: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		var rsp = Get(t, frontend, BuildPath(test.path))
		if rsp.StatusCode != test.expect {
			t.Fatalf("访问：%s 期望: %d，实际：%d \n", BuildPath(test.path), test.expect, rsp.StatusCode)
		}
	}
}

func TestEngine_Abort(t *testing.T) {
	var backend = NewBackend()
	defer backend.Close()

	var engine = rx.New()
	engine.Load(NewProvider(backend.URL))

	// 将所有请求的返回状态码全部调整为 http.StatusUnauthorized
	engine.Use(func(c *rx.Context) {
		c.AbortWithStatus(http.StatusUnauthorized)
	})

	engine.Use(func(c *rx.Context) {
		t.Fatal("不应该执行到这里")
	})

	engine.NoRoute(func(c *rx.Context) {
		t.Fatal("不应该执行到这里")
	})

	frontend := httptest.NewServer(engine)
	defer frontend.Close()

	var tests = []struct {
		path   int
		expect int
	}{
		{
			path:   http.StatusOK,
			expect: http.StatusUnauthorized,
		},
		{
			path:   http.StatusBadRequest,
			expect: http.StatusUnauthorized,
		},
		{
			path:   http.StatusBadGateway,
			expect: http.StatusUnauthorized,
		},
	}

	for _, test := range tests {
		var rsp = Get(t, frontend, BuildPath(test.path))
		if rsp.StatusCode != test.expect {
			t.Fatalf("访问：%s 期望: %d，实际：%d \n", BuildPath(test.path), test.expect, rsp.StatusCode)
		}
	}
}

func TestEngine_Error(t *testing.T) {
	var engine = rx.New()
	engine.Load(NewProvider("http://127.0.0.1:1100"))

	engine.HandleError(func(c *rx.Context, err error) {
		switch c.Request.URL.Path {
		case BuildPath(http.StatusOK):
			c.AbortWithStatus(http.StatusNotImplemented)
		case BuildPath(http.StatusBadRequest):
			c.AbortWithStatus(http.StatusServiceUnavailable)
		}
	})

	frontend := httptest.NewServer(engine)
	defer frontend.Close()

	var tests = []struct {
		path   int
		expect int
	}{
		{
			path:   http.StatusOK,
			expect: http.StatusNotImplemented,
		},
		{
			path:   2001,
			expect: http.StatusInternalServerError,
		},
		{
			path:   http.StatusBadRequest,
			expect: http.StatusServiceUnavailable,
		},
		{
			path:   http.StatusBadGateway, // provider 中没有注册该路由规则，所以返回 http.StatusBadGateway
			expect: http.StatusBadGateway,
		},
	}

	for _, test := range tests {
		var rsp = Get(t, frontend, BuildPath(test.path))
		if rsp.StatusCode != test.expect {
			t.Fatalf("访问：%s 期望: %d，实际：%d \n", BuildPath(test.path), test.expect, rsp.StatusCode)
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

func TestEngine_D1(t *testing.T) {
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
			path:   "/user/list",
			expect: http.StatusOK,
		},
		{
			path:   "/user/lists",
			expect: http.StatusBadRequest,
		},
		{
			path:   "/user/123",
			expect: http.StatusOK,
		},
		{
			path:   "/user/456",
			expect: http.StatusOK,
		},
		{
			path:   "/user/789",
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
			path:   "/book/list",
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

func TestEngine_Rewrite(t *testing.T) {
	var userBackend = NewUserBackend()
	defer userBackend.Close()

	var provider = rx.NewListProvider()
	provider.Add("/api/user", []string{userBackend.URL})

	var engine = rx.New()
	engine.Load(provider)

	engine.Use(func(c *rx.Context) {
		// 删除 URL Path 中的 /api
		c.Request.URL.Path = strings.TrimPrefix(c.Request.URL.Path, "/api")
	})

	frontend := httptest.NewServer(engine)
	defer frontend.Close()

	var tests = []struct {
		path   string
		expect int
	}{
		{
			path:   "/api/user/list",
			expect: http.StatusOK,
		},
		{
			path:   "/api/user/lists",
			expect: http.StatusBadRequest,
		},
		{
			path:   "/api/user/123",
			expect: http.StatusOK,
		},
		{
			path:   "/api/user/456",
			expect: http.StatusOK,
		},
		{
			path:   "/api/user/789",
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
