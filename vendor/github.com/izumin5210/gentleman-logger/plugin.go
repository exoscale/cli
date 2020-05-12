package logger

import (
	"io"
	"net/http"

	"github.com/izumin5210/httplogger"
	"gopkg.in/h2non/gentleman.v2/context"
	"gopkg.in/h2non/gentleman.v2/plugin"
)

// New creates logger plugin instance
func New(out io.Writer) plugin.Plugin {
	return new(func(parent http.RoundTripper) http.RoundTripper {
		return httplogger.NewRoundTripper(out, parent)
	})
}

// FromLogger creates logger plugin instance with a specified logger implementation
func FromLogger(writer httplogger.SimpleLogWriter) plugin.Plugin {
	return new(func(parent http.RoundTripper) http.RoundTripper {
		return httplogger.FromSimpleLogger(writer, parent)
	})
}

func new(transportFn func(parent http.RoundTripper) http.RoundTripper) plugin.Plugin {
	return plugin.NewRequestPlugin(func(c *context.Context, h context.Handler) {
		c.Client.Transport = transportFn(c.Client.Transport)
		h.Next(c)
	})
}
