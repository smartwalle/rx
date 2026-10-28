package rx

import (
	"io"
	"net/http"
)

var kContentTypeText = []string{"text/plain; charset=utf-8"}

type textRender struct {
	text string
}

func (r textRender) ContentType() []string {
	return kContentTypeText
}

func (r textRender) Render(w http.ResponseWriter) error {
	_, err := io.WriteString(w, r.text)
	return err
}

func (r textRender) WriteContentType(w http.ResponseWriter) {
	var header = w.Header()
	if val := header[kContentType]; len(val) == 0 {
		header[kContentType] = r.ContentType()
	}
}
