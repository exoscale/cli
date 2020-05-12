package httplogger

import (
	"net/http"
	"time"
)

// RequestLog contains http(s) request information
type RequestLog struct {
	*http.Request
	RequestedAt time.Time
}

// ResponseLog contains http(s) response information or errors
type ResponseLog struct {
	*http.Response
	DurationNano int64
	Error        error
}

// SimpleLogWriter is interface for writing logs
type SimpleLogWriter interface {
	Print(v ...interface{})
}

// LogWriter is interface for writing logs
type LogWriter interface {
	PrintRequest(req *RequestLog)
	PrintResponse(resp *ResponseLog)
}
