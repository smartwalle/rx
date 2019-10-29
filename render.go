package rx

import "net/http"

type Render interface {
	Render(http.ResponseWriter) error

	ContentType() []string
}
