package utils

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// WarningSink buffers per-zone warnings so they don't interleave with
// streamed stdout output. Flush is called explicitly (via defer) at the
// end of the command, or implicitly on SIGINT/SIGTERM via
// InstallSignalFlush.
type WarningSink struct {
	mu   sync.Mutex
	msgs []string
	out  io.Writer
}

// NewWarningSink returns a sink that writes to os.Stderr on Flush.
func NewWarningSink() *WarningSink {
	return &WarningSink{out: os.Stderr}
}

// NewWarningSinkTo returns a sink that writes to w on Flush. Used in tests.
func NewWarningSinkTo(w io.Writer) *WarningSink {
	return &WarningSink{out: w}
}

// Add buffers a formatted warning. Safe to call from multiple goroutines.
func (s *WarningSink) Add(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	s.mu.Lock()
	s.msgs = append(s.msgs, msg)
	s.mu.Unlock()
}

// Len returns the number of buffered warnings.
func (s *WarningSink) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.msgs)
}

// Flush writes every buffered warning to the sink's writer and clears
// the buffer. Safe to call multiple times.
func (s *WarningSink) Flush() {
	s.mu.Lock()
	msgs := s.msgs
	s.msgs = nil
	s.mu.Unlock()
	for _, m := range msgs {
		fmt.Fprintf(s.out, "warning: %s\n", m)
	}
}
