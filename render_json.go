package rx

import (
	"encoding/json"
	"net/http"
)

var kContentTypeJSON = []string{"application/json; charset=utf-8"}

type jsonRender struct {
	data interface{}
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
