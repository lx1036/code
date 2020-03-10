package kube_gin

import (
	"io"
	"log"
	"time"
)
func Recovery() HandlerFunc {
	return RecoveryWithWriter(DefaultErrorWriter)
}

// RecoveryWithWriter returns a middleware for a given writer that recovers from any panics and writes a 500 if there was one.
func RecoveryWithWriter(out io.Writer) HandlerFunc {
	var logger *log.Logger
	if out != nil {
		logger = log.New(out, "\n\n\x1b[31m", log.LstdFlags)
	}

	return func(context *Context) {
		defer func() {
			if err := recover(); err != nil {
				if logger != nil {
					logger.Printf("[Recovery] %s panic recovered:\n%s", timeFormat(time.Now()), err)
				}
			}
		}()

		context.Next()
	}
}

func timeFormat(t time.Time) string {
	var timeString = t.Format("2006/01/02 - 15:04:05")
	return timeString
}
