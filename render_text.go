package rx

import (
	"fmt"
	"net/http"
)

var kContentTypeText = []string{"text/plain; charset=utf-8"}

type TextRender struct {
	text string
}

func (r TextRender) ContentType() []string {
	return kContentTypeText
}

func (r TextRender) Render(w http.ResponseWriter) error {
	_, err := fmt.Fprintf(w, r.text)
	return err
}
