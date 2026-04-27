package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
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

// Default minimum content width when neither the outputWidth tag nor the
// header label hints at something better. Wide enough for short status
// codes; long values just expand the cell.
const defaultTableMinWidth = 12

type tableStreamer struct {
	mu      sync.Mutex
	w       io.Writer
	headers []string
	widths  []int // content widths; cell visual width is widths[i]+2
	started bool
	closed  bool
}

func newTableStreamer(rowType any, w io.Writer) *tableStreamer {
	t := rowKindOf(rowType)
	headers := outputTableHeaders(t)
	widths := tableColumnWidths(t, headers)
	// Match tablewriter's auto-format: uppercase headers.
	for i := range headers {
		headers[i] = strings.ToUpper(headers[i])
	}
	return &tableStreamer{w: w, headers: headers, widths: widths}
}

// tableColumnWidths derives a content width per column from the
// outputWidth struct tag, falling back to max(header length,
// defaultTableMinWidth).
func tableColumnWidths(t reflect.Type, headers []string) []int {
	widths := make([]int, len(headers))
	col := 0
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if l, ok := f.Tag.Lookup("output"); ok && l == "-" {
			continue
		}
		if w, ok := f.Tag.Lookup("outputWidth"); ok {
			if n, err := strconv.Atoi(w); err == nil && n > 0 {
				widths[col] = n
				col++
				continue
			}
		}
		hw := len(headers[col])
		if hw > defaultTableMinWidth {
			widths[col] = hw
		} else {
			widths[col] = defaultTableMinWidth
		}
		col++
	}
	return widths
}

func (s *tableStreamer) Push(row any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.started {
		s.writeBorder()
		s.writeRow(s.headers)
		s.writeBorder()
		s.started = true
	}
	cells := outputTableRow(reflect.Indirect(reflect.ValueOf(row)))
	s.writeRow(cells)
	return nil
}

// writeBorder writes a "┼───┼───┼" line spanning all columns.
func (s *tableStreamer) writeBorder() {
	var b strings.Builder
	b.WriteString("┼")
	for _, w := range s.widths {
		b.WriteString(strings.Repeat("─", w+2))
		b.WriteString("┼")
	}
	b.WriteString("\n")
	_, _ = s.w.Write([]byte(b.String()))
}

// writeRow writes "│ cell1 │ cell2 │" with one-space padding on each
// side. Cells longer than the configured width expand the cell rather
// than wrap or truncate, mirroring SetAutoWrapText(false) on the
// existing table.
func (s *tableStreamer) writeRow(cells []string) {
	var b strings.Builder
	b.WriteString("│")
	for i, c := range cells {
		w := 0
		if i < len(s.widths) {
			w = s.widths[i]
		}
		// Pad right to at least w; if c is wider, the cell expands.
		fmt.Fprintf(&b, " %-*s ", w, c)
		b.WriteString("│")
	}
	b.WriteString("\n")
	_, _ = s.w.Write([]byte(b.String()))
}

func (s *tableStreamer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	if s.started {
		s.writeBorder()
	}
	return nil
}

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
