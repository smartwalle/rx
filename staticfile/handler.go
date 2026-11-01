package staticfile

import (
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/smartwalle/rx"
)

// NewHandler 返回一个静态文件处理器。
//
// root 是静态文件根目录，可以是相对路径或绝对路径。
// 相对路径在创建 Handler 时相对于当前工作目录解析。
//
// 仅支持 GET 和 HEAD 请求。使用时需要分别注册 GET 和 HEAD 路由：
//
//	handler := staticfile.NewHandler("./public")
//	r.GET("/static/*filepath", handler)
//	r.HEAD("/static/*filepath", handler)
//
// 路由中的通配符参数必须命名为 filepath，Handler 通过
// c.Param("filepath") 获取相对于 root 的文件路径。
// 例如：
//
//	请求路径：/static/css/app.css
//	路由定义：/static/*filepath
//	参数值：  css/app.css
//	对应文件：./public/css/app.css
//
// 只提供普通文件，不提供目录列表，也不默认查找 index.html。
// 文件不存在、路径非法或请求目标不是普通文件时返回 404。
//
// root 必须是可信的静态文件目录，不应允许不可信用户在服务运行期间
// 修改目录内容。
//
// 如果 root 为空、无法解析或不是目录，会 panic。
func NewHandler(root string) rx.HandlerFunc {
	if root == "" {
		panic("root must not be empty")
	}

	root, err := filepath.Abs(root)
	if err != nil {
		panic(err)
	}

	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		panic(err)
	}

	info, err := os.Stat(root)
	if err != nil {
		panic(err)
	}

	if !info.IsDir() {
		panic("root is not a directory")
	}

	return func(c *rx.Context) {
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.Writer.Header().Set("Allow", "GET, HEAD")
			http.Error(c.Writer, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var name = c.Param("filepath")
		if strings.HasPrefix(name, "//") {
			http.NotFound(c.Writer, c.Request)
			return
		}
		name = strings.TrimPrefix(name, "/")
		if strings.Contains(name, "\\") ||
			!fs.ValidPath(name) ||
			!filepath.IsLocal(filepath.FromSlash(name)) {
			http.NotFound(c.Writer, c.Request)
			return
		}

		var file *os.File
		file, info, err = open(root, name)
		if err != nil {
			http.NotFound(c.Writer, c.Request)
			return
		}
		defer file.Close()

		http.ServeContent(c.Writer, c.Request, path.Base(name), info.ModTime(), file)
	}
}

// open 打开 root 下的普通文件，并检查符号链接解析后的路径是否仍在 root 内。
func open(root, name string) (*os.File, os.FileInfo, error) {
	var filename = filepath.Join(root, filepath.FromSlash(name))

	resolved, err := filepath.EvalSymlinks(filename)
	if err != nil {
		return nil, nil, err
	}

	rel, err := filepath.Rel(root, resolved)
	if err != nil {
		return nil, nil, err
	}

	if !filepath.IsLocal(rel) || rel == "." {
		return nil, nil, os.ErrPermission
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return nil, nil, err
	}

	if !info.Mode().IsRegular() {
		return nil, nil, os.ErrPermission
	}

	file, err := openFile(resolved)
	if err != nil {
		return nil, nil, err
	}

	openedInfo, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, nil, err
	}

	if !openedInfo.Mode().IsRegular() || !os.SameFile(info, openedInfo) {
		file.Close()
		return nil, nil, os.ErrPermission
	}

	return file, openedInfo, nil
}
