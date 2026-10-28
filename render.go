package rx

import "net/http"

type Render interface {
	ContentType() []string

	Render(http.ResponseWriter) error

	WriteContentType(http.ResponseWriter)
}
