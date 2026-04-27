package utils

import (
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/term"

	"github.com/exoscale/cli/pkg/globalstate"
)

// spinnerFrames are simple braille glyphs that render well in any
// modern terminal.
var spinnerFrames = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}

// Spinner draws a single rotating glyph at column 0 of stderr until
// Stop is called. By design the spinner occupies exactly one
// character, so concurrent writes to stdout (which start at column 0
// after a newline) cleanly overwrite the glyph — no caption ghosts
// past the leading '│' of a streamed table row.
//
// Silent if stderr isn't a TTY or globalstate.Quiet is on, so
// callers can use Start/Stop unconditionally.
type Spinner struct {
	w      io.Writer
	stop   chan struct{}
	done   chan struct{}
	active bool
}

func NewSpinner() *Spinner {
	return &Spinner{
		w:    os.Stderr,
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
}

// SetWriter overrides the writer the spinner draws on. Must be called
// before Start. If w is not an *os.File pointing at a TTY, Start is
// a no-op.
func (s *Spinner) SetWriter(w io.Writer) {
	s.w = w
}

func (s *Spinner) Start() {
	if globalstate.Quiet {
		return
	}
	f, ok := s.w.(*os.File)
	if !ok || !term.IsTerminal(int(f.Fd())) {
		return
	}
	s.active = true
	go s.run()
}

// Stop erases the glyph and returns once the drawing goroutine has
// exited. Safe to call multiple times and safe to call when Start was
// a no-op.
func (s *Spinner) Stop() {
	if !s.active {
		return
	}
	s.active = false
	close(s.stop)
	<-s.done
}

func (s *Spinner) run() {
	defer close(s.done)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	frame := 0
	draw := func() {
		// Single char + \r: occupies col 0, cursor returns to col 0.
		// Any subsequent stdout byte at col 0 overwrites the glyph
		// cleanly, so streamed rows don't leave caption ghosts.
		fmt.Fprintf(s.w, "%c\r", spinnerFrames[frame%len(spinnerFrames)])
	}
	draw()
	for {
		select {
		case <-s.stop:
			// Wipe col 0 with a space, then return to col 0.
			fmt.Fprint(s.w, " \r")
			return
		case <-ticker.C:
			frame++
			draw()
		}
	}
}
