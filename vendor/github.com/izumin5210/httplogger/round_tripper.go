package httplogger

import (
	"io"
	"net/http"
	"time"
)

type loggingTransport struct {
	writer LogWriter
	parent http.RoundTripper
}

// NewRoundTripper returns new RoundTripper instance for logging http request and response
func NewRoundTripper(out io.Writer, parent http.RoundTripper) http.RoundTripper {
	return &loggingTransport{
		writer: newDefaultLogWriter(out),
		parent: parent,
	}
}

// FromSimpleLogger creates new logging RoundTripper instance with given log writer
func FromSimpleLogger(writer SimpleLogWriter, parent http.RoundTripper) http.RoundTripper {
	return &loggingTransport{
		writer: wrapSimpleLogWriter(writer),
		parent: parent,
	}
}

func (lt *loggingTransport) parentTransport() http.RoundTripper {
	if lt.parent == nil {
		return http.DefaultTransport
	}
	return lt.parent
}

func (lt *loggingTransport) CancelRequest(req *http.Request) {
	type canceler interface {
		CancelRequest(*http.Request)
	}
	if cr, ok := lt.parentTransport().(canceler); ok {
		cr.CancelRequest(req)
	}
}

func (lt *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	requestedAt := time.Now()
	lt.writer.PrintRequest(&RequestLog{Request: req, RequestedAt: requestedAt})

	resp, err := lt.parentTransport().RoundTrip(req)

	respTime := time.Since(requestedAt)
	lt.writer.PrintResponse(&ResponseLog{Response: resp, DurationNano: respTime.Nanoseconds(), Error: err})

	return resp, err
}
