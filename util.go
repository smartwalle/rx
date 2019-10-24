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
	if fistChar(path) == '/' {
		return ps[1:]
	}
	return ps
}

func cleanPath(path string) string {
	if len(path) > 1 && lastChar(path) == '/' {
		return strings.TrimSuffix(path, "/")
	}
	return path
}

func lastChar(str string) uint8 {
	asset(str != "", "the length of the string can't be 0")
	return str[len(str)-1]
}

func fistChar(str string) uint8 {
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
