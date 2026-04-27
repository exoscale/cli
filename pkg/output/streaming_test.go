package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/exoscale/cli/pkg/globalstate"
)

type streamRow struct {
	Name string `json:"name"`
	Zone string `json:"zone"`
	N    int    `json:"n"`
}

func withFormat(t *testing.T, f string) func() {
	t.Helper()
	prev := globalstate.OutputFormat
	prevTpl := GOutputTemplate
	globalstate.OutputFormat = f
	GOutputTemplate = ""
	return func() {
		globalstate.OutputFormat = prev
		GOutputTemplate = prevTpl
	}
}

func TestStreamingJSON(t *testing.T) {
	defer withFormat(t, "json")()
	var buf bytes.Buffer
	s := NewStreamer(streamRow{}, &buf)

	rows := []streamRow{
		{Name: "a", Zone: "z1", N: 1},
		{Name: "b", Zone: "z2", N: 2},
		{Name: "c", Zone: "z3", N: 3},
	}
	for _, r := range rows {
		if err := s.Push(r); err != nil {
			t.Fatalf("push: %v", err)
		}
	}
	if err := s.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	var got []streamRow
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %v\nraw: %s", err, buf.String())
	}
	if len(got) != 3 {
		t.Fatalf("want 3 rows, got %d", len(got))
	}
	for i, r := range rows {
		if got[i] != r {
			t.Errorf("row %d: got %+v want %+v", i, got[i], r)
		}
	}
}

func TestStreamingJSONEmpty(t *testing.T) {
	defer withFormat(t, "json")()
	var buf bytes.Buffer
	s := NewStreamer(streamRow{}, &buf)
	if err := s.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	got := strings.TrimSpace(buf.String())
	if got != "[]" {
		t.Fatalf("empty stream: want '[]', got %q", got)
	}
}

func TestStreamingText(t *testing.T) {
	defer withFormat(t, "text")()
	var buf bytes.Buffer
	s := NewStreamer(streamRow{}, &buf)

	if err := s.Push(streamRow{Name: "a", Zone: "z1", N: 1}); err != nil {
		t.Fatalf("push: %v", err)
	}
	if err := s.Push(streamRow{Name: "b", Zone: "z2", N: 2}); err != nil {
		t.Fatalf("push: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 lines, got %d: %q", len(lines), buf.String())
	}
	if !strings.Contains(lines[0], "a") || !strings.Contains(lines[0], "z1") {
		t.Errorf("line 0 missing fields: %q", lines[0])
	}
	if !strings.Contains(lines[1], "b") || !strings.Contains(lines[1], "z2") {
		t.Errorf("line 1 missing fields: %q", lines[1])
	}
}

func TestStreamingTable(t *testing.T) {
	defer withFormat(t, "")() // default = table
	var buf bytes.Buffer
	s := NewStreamer(streamRow{}, &buf)

	if err := s.Push(streamRow{Name: "a", Zone: "z1", N: 1}); err != nil {
		t.Fatalf("push: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	out := buf.String()
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	// Expected: top border, header, header-sep, row, bottom border = 5 lines.
	if len(lines) != 5 {
		t.Fatalf("want 5 lines (top, header, sep, row, bottom), got %d:\n%s", len(lines), out)
	}
	for _, want := range []string{"NAME", "ZONE", "N"} {
		if !strings.Contains(lines[1], want) {
			t.Errorf("header missing %q: %q", want, lines[1])
		}
	}
	if !strings.Contains(lines[3], "a") || !strings.Contains(lines[3], "z1") {
		t.Errorf("row missing fields: %q", lines[3])
	}
	for _, idx := range []int{0, 2, 4} {
		if !strings.HasPrefix(lines[idx], "┼") {
			t.Errorf("line %d should start with ┼, got %q", idx, lines[idx])
		}
	}
	if !strings.HasPrefix(lines[1], "│") || !strings.HasPrefix(lines[3], "│") {
		t.Errorf("header/row should start with │")
	}
}

func TestStreamingTableEmpty(t *testing.T) {
	defer withFormat(t, "")()
	var buf bytes.Buffer
	s := NewStreamer(streamRow{}, &buf)
	if err := s.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("empty stream should produce no output, got %q", buf.String())
	}
}

func TestStreamingTable_OutputWidthTag(t *testing.T) {
	defer withFormat(t, "")()
	type narrowRow struct {
		ID   string `outputWidth:"4"`
		Name string `outputWidth:"6"`
	}
	var buf bytes.Buffer
	s := NewStreamer(narrowRow{}, &buf)
	if err := s.Push(narrowRow{ID: "x", Name: "y"}); err != nil {
		t.Fatalf("push: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	// Border with widths 4 and 6 → content runs of 6 and 8 dashes
	// (width + 2 padding spaces).
	wantBorder := "┼" + strings.Repeat("─", 6) + "┼" + strings.Repeat("─", 8) + "┼"
	if !strings.Contains(buf.String(), wantBorder) {
		t.Errorf("expected border %q in output:\n%s", wantBorder, buf.String())
	}
}

func TestStreamingConcurrentPush(t *testing.T) {
	defer withFormat(t, "json")()
	var buf bytes.Buffer
	s := NewStreamer(streamRow{}, &buf)

	const N = 50
	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func(i int) {
			defer wg.Done()
			_ = s.Push(streamRow{Name: fmt.Sprintf("n%d", i), Zone: "z", N: i})
		}(i)
	}
	wg.Wait()
	if err := s.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	var got []streamRow
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(got) != N {
		t.Fatalf("want %d rows, got %d", N, len(got))
	}
}

func TestStreamingJSONRowTypeMismatch(t *testing.T) {
	defer withFormat(t, "json")()
	var buf bytes.Buffer
	s := NewStreamer(streamRow{}, &buf)
	type other struct{ X int }
	if err := s.Push(other{X: 1}); err == nil {
		t.Fatal("want type mismatch error")
	}
}
