package rx

import (
	"encoding/json"
	"net/http"
)

var kContentTypeJSON = []string{"application/json; charset=utf-8"}

type jsonRender struct {
	data any
}

func (r jsonRender) ContentType() []string {
	return kContentTypeJSON
}

func (r jsonRender) Render(w http.ResponseWriter) error {
	bytes, err := json.Marshal(r.data)
	if err != nil {
		return err
	}
	_, err = w.Write(bytes)
	return err
}

func (r jsonRender) WriteContentType(w http.ResponseWriter) {
	var header = w.Header()
	if val := header[kContentType]; len(val) == 0 {
		header[kContentType] = r.ContentType()
	}
}
