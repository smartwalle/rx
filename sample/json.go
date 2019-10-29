package main

import (
	"encoding/json"
	"github.com/smartwalle/rx"
	"net/http"
)

func main() {
	var s = rx.New()
	s.GET("/json", func(c *rx.Context) {
		var r = JSONRender{data: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}}
		c.Render(http.StatusOK, r)
	})
	http.ListenAndServe(":8897", s)
}

var contentType = []string{"application/json; charset=utf-8"}

type JSONRender struct {
	data interface{}
}

func (this JSONRender) ContentType() []string {
	return contentType
}

func (this JSONRender) Render(w http.ResponseWriter) error {
	bytes, err := json.Marshal(this.data)
	if err != nil {
		return err
	}
	_, err = w.Write(bytes)
	return err
}
