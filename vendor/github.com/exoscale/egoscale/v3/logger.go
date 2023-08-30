package v3

import (
	"fmt"
	"io"
)

// Logger is an interface that can be implemented by to provide logging for Client.
// Interface is the same as LeveledLogger from hashicorp/go-retryablehttp (default http client).
// Argument keysAndValues expects even number of values and each odd item a string.
type Logger interface {
	Error(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
}

// StandardLogger is a simple logger that implements Logger interface.
// It supports any output that implements Writer interface.
// Set Debug to true to print debug messages.
type StandardLogger struct {
	debug  bool
	writer io.Writer
}

func NewStandardLogger(writer io.Writer, debug bool) *StandardLogger {
	return &StandardLogger{debug: debug, writer: writer}
}

func (l *StandardLogger) Error(msg string, keysAndValues ...interface{}) {
	fmt.Fprintf(l.writer, "[ERROR] %s %v\n", msg, l.toMap(keysAndValues))
}

func (l *StandardLogger) Info(msg string, keysAndValues ...interface{}) {
	fmt.Fprintf(l.writer, "[INFO] %s %v\n", msg, l.toMap(keysAndValues))
}

func (l *StandardLogger) Debug(msg string, keysAndValues ...interface{}) {
	if l.debug {
		fmt.Fprintf(l.writer, "[DEBUG] %s %v\n", msg, l.toMap(keysAndValues))
	}
}

func (l *StandardLogger) Warn(msg string, keysAndValues ...interface{}) {
	fmt.Fprintf(l.writer, "[WARN] %s %v\n", msg, l.toMap(keysAndValues))
}

// Convert keysAndValues to map.
// Every odd element (key) must be a string.
// If slice has odd number of elements, last item with value nil is added.
func (l *StandardLogger) toMap(keysAndValues []interface{}) map[string]interface{} {
	if len(keysAndValues) == 0 {
		return nil
	}

	if len(keysAndValues)%2 > 0 {
		keysAndValues = append(keysAndValues, nil)
	}

	m := map[string]interface{}{}
	for i := 0; i < len(keysAndValues); i = i + 2 {
		if k, ok := keysAndValues[i].(string); ok {
			m[k] = keysAndValues[i+1]
		} else {
			l.Warn("standard_logger: key is not a string", "key", keysAndValues[i], "val", keysAndValues[i+1])
		}
	}

	return m
}
