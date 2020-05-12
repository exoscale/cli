package httplogger

import (
	"fmt"
	"io"
	"log"
	"net/http/httputil"
	"strings"
)

const (
	defaultPrefix = "[http] "
)

type defaultLogWriter struct {
	writer SimpleLogWriter
}

func newDefaultLogWriter(out io.Writer) LogWriter {
	return wrapSimpleLogWriter(log.New(out, defaultPrefix, log.LstdFlags))
}

func wrapSimpleLogWriter(writer SimpleLogWriter) LogWriter {
	return &defaultLogWriter{
		writer: writer,
	}
}

func (l *defaultLogWriter) PrintRequest(reqLog *RequestLog) {
	dump, _ := httputil.DumpRequest(reqLog.Request, true)
	l.writer.Print(fmt.Sprintf("--> %s\n", strings.Replace(string(dump), "\r\n", "\n", -1)))
}

func (l *defaultLogWriter) PrintResponse(respLog *ResponseLog) {
	if respLog.Response == nil {
		l.writer.Print(fmt.Sprintf("--> %s (%dms)\n", respLog.Error.Error(), respLog.DurationNano/1e6))
		return
	}

	dump, _ := httputil.DumpResponse(respLog.Response, true)
	lines := strings.Split(string(dump), "\r\n")
	lines[0] = fmt.Sprintf("<-- %s (%dms)", lines[0], respLog.DurationNano/1e6)
	l.writer.Print(strings.Join(lines, "\n") + "\n")
}
