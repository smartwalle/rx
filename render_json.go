package rx

import (
	"encoding/json"
	"net/http"
)

var kContentTypeJSON = []string{"application/json; charset=utf-8"}

type JSONRender struct {
	data interface{}
}

func (this JSONRender) ContentType() []string {
	return kContentTypeJSON
}

func (this JSONRender) Render(w http.ResponseWriter) error {
	bytes, err := json.Marshal(this.data)
	if err != nil {
		return err
	}
	_, err = w.Write(bytes)
	return err
}
