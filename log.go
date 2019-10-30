package rx

import (
	"fmt"
	"log"
	"os"
	"time"
)

func init() {
	SetLogger(log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile))
}

type Logger interface {
	SetPrefix(prefix string)
	Prefix() string
	Println(args ...interface{})
	Printf(format string, args ...interface{})
	Output(calldepth int, s string) error
}

var logger Logger

func SetLogger(l Logger) {
	if l == nil {
		l = &nilLogger{}
	}
	if l.Prefix() == "" {
		l.SetPrefix("[RX] ")
	}
	logger = l
}

type nilLogger struct {
}

func (this *nilLogger) SetPrefix(prefix string) {
}

func (this *nilLogger) Prefix() string {
	return ""
}

func (this *nilLogger) Println(args ...interface{}) {
}

func (this *nilLogger) Printf(format string, args ...interface{}) {
}

func (this *nilLogger) Output(calldepth int, s string) error {
	return nil
}

func Log() HandlerFunc {
	return func(c *Context) {
		var beginTime = time.Now()
		c.Next()
		var endTime = time.Now()

		var duration = endTime.Sub(beginTime)
		var writer = c.Writer

		var method = c.Request.Method
		var path = c.Request.URL.Path
		var rawQuery = c.Request.URL.RawQuery
		if rawQuery != "" {
			path = path + "?" + rawQuery
		}

		logger.Output(1, fmt.Sprintf("| %d | %10s | %8s - %s", writer.StatusCode(), duration, method, path))
	}
}
