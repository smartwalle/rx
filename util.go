package rx

import (
	"net/http"
	"path"
	"reflect"
	"runtime"
	"strings"
)

func asset(guard bool, text string) {
	if !guard {
		panic(text)
	}
}

func splitPath(path string) []string {
	path = strings.TrimPrefix(path, "/")
	var ps = strings.Split(path, "/")
	return ps
}

func CleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	//p = path.Clean(p)
	return p
}

func lastChar(str string) uint8 {
	asset(str != "", "the length of the string can't be 0")
	return str[len(str)-1]
}

func firstChar(str string) uint8 {
	asset(str != "", "the length of the string can't be 0")
	return str[0]
}

func joinPaths(absolutePath, relativePath string) string {
	if relativePath == "" {
		return absolutePath
	}

	finalPath := path.Join(absolutePath, relativePath)
	appendSlash := lastChar(relativePath) == '/' && lastChar(finalPath) != '/'
	if appendSlash {
		return finalPath + "/"
	}
	return finalPath
}

func nameOfFunction(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func bodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == http.StatusNoContent:
		return false
	case status == http.StatusNotModified:
		return false
	}
	return true
}
