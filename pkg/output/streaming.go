package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"
	"text/template"

	"github.com/exoscale/cli/pkg/globalstate"
)

// StreamingOutputter is implemented by row-typed outputs that can be
// emitted progressively. Push is goroutine-safe; Close finalizes the
// render (flushes JSON, etc.).
type StreamingOutputter interface {
	Push(row any) error
	Close() error
}

// NewStreamer returns the right StreamingOutputter for the active
// globalstate.OutputFormat. rowType is a zero-value of the row struct
// used to derive table headers and the JSON envelope element type.
// w is where the streamer writes (typically os.Stdout).
func NewStreamer(rowType any, w io.Writer) StreamingOutputter {
	if GOutputTemplate != "" {
		return newTextStreamer(rowType, w, GOutputTemplate)
	}
	switch globalstate.OutputFormat {
	case "json":
		return newJSONStreamer(rowType, w)
	case "text":
		return newTextStreamer(rowType, w, "")
	default:
		return newTableStreamer(rowType, w)
	}
}

// rowKindOf returns the underlying struct type of rowType.
func rowKindOf(rowType any) reflect.Type {
	t := reflect.TypeOf(rowType)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// --- text ---

type textStreamer struct {
	mu  sync.Mutex
	w   io.Writer
	tpl *template.Template
}

func newTextStreamer(rowType any, w io.Writer, userTpl string) *textStreamer {
	tpl := userTpl
	if tpl == "" {
		zero := reflect.New(rowKindOf(rowType)).Elem().Interface()
		fields := TemplateAnnotations(zero)
		for i := range fields {
			fields[i] = "{{" + fields[i] + "}}"
		}
		tpl = strings.Join(fields, "\t")
	}
	t, err := template.New("out").Parse(tpl)
	if err != nil {
		// Fall back to a printf so we still produce *something* and
		// surface the misconfiguration rather than crashing the CLI.
		t = template.Must(template.New("out").Parse("{{.}}"))
		fmt.Fprintf(w, "warning: invalid output template: %s\n", err)
	}
	return &textStreamer{w: w, tpl: t}
}

func (s *textStreamer) Push(row any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.tpl.Execute(s.w, row); err != nil {
		return fmt.Errorf("text stream: %w", err)
	}
	_, err := fmt.Fprintln(s.w)
	return err
}

func (s *textStreamer) Close() error { return nil }

// --- table ---

type tableStreamer struct {
	mu      sync.Mutex
	w       io.Writer
	headers []string
	widths  []int
	started bool
}

func newTableStreamer(rowType any, w io.Writer) *tableStreamer {
	headers := outputTableHeaders(rowKindOf(rowType))
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	return &tableStreamer{w: w, headers: headers, widths: widths}
}

func (s *tableStreamer) Push(row any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.started {
		s.writeHeader()
		s.started = true
	}
	cells := outputTableRow(reflect.Indirect(reflect.ValueOf(row)))
	s.writeRow(cells)
	return nil
}

func (s *tableStreamer) writeHeader() {
	s.writeRow(s.headers)
	sep := make([]string, len(s.headers))
	for i, w := range s.widths {
		sep[i] = strings.Repeat("─", w)
	}
	s.writeRow(sep)
}

func (s *tableStreamer) writeRow(cells []string) {
	parts := make([]string, len(cells))
	for i, c := range cells {
		w := 0
		if i < len(s.widths) {
			w = s.widths[i]
		}
		parts[i] = fmt.Sprintf("%-*s", w, c)
	}
	fmt.Fprintln(s.w, strings.Join(parts, "  "))
}

func (s *tableStreamer) Close() error { return nil }

// --- json ---

type jsonStreamer struct {
	mu      sync.Mutex
	w       io.Writer
	rowType reflect.Type
	rows    reflect.Value // []rowType
	closed  bool
}

func newJSONStreamer(rowType any, w io.Writer) *jsonStreamer {
	rt := rowKindOf(rowType)
	sliceType := reflect.SliceOf(rt)
	return &jsonStreamer{
		w:       w,
		rowType: rt,
		rows:    reflect.MakeSlice(sliceType, 0, 0),
	}
}

func (s *jsonStreamer) Push(row any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	v := reflect.Indirect(reflect.ValueOf(row))
	if v.Type() != s.rowType {
		return fmt.Errorf("json stream: row type mismatch: got %s want %s", v.Type(), s.rowType)
	}
	s.rows = reflect.Append(s.rows, v)
	return nil
}

func (s *jsonStreamer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	enc := json.NewEncoder(s.w)
	enc.SetEscapeHTML(false)
	return enc.Encode(s.rows.Interface())
}
