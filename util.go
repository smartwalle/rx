package rx

import (
	"path"
	"strings"
)

func asset(guard bool, text string) {
	if !guard {
		panic(text)
	}
}

func splitPath(path string) []string {
	var ps = strings.Split(path, "/")
	if firstChar(path) == '/' {
		return ps[1:]
	}
	return ps
}

func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	p = path.Clean(p)
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
