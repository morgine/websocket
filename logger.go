package websocket

import (
	"github.com/morgine/log"
	"io"
	"time"
)

func New1Logger() Handler {
	var h HandlerFunc = func(c *Context) {
		lw := &logWriter{
			Writer: c.Writer,
		}
		c.Writer = lw
		start := time.Now()
		defer func() {
			end := time.Now()
			latency := end.Sub(start)
			clientIP := c.ClientIP()
			request := c.GetRequestBody()
			response := lw.data
			log.Info.Printf("[WS] %v | %13v | %s | request:\n%s\nresponse:\n%s\n",
				end.Format("2006/01/02 - 15:04:05"),
				latency,
				clientIP,
				request,
				response,
			)
		}()
		c.Next()
	}
	return h
}

type logWriter struct {
	data []byte
	io.Writer
}

func (l *logWriter) Write(p []byte) (n int, err error) {
	l.data = p
	return l.Writer.Write(p)
}
