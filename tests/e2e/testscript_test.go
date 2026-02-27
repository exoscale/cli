package e2e_test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/creack/pty"
	"github.com/rogpeppe/go-internal/testscript"
)

var (
	exoBinary string
	cliRoot   string // Set at init time before working directory changes
)

func init() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get source file location")
	}
	cliRoot = filepath.Dir(filepath.Dir(filepath.Dir(filename)))
	exoBinary = filepath.Join(cliRoot, "bin", "exo")
}

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"exo": mainExo,
	})
}

func mainExo() {
	if _, err := os.Stat(exoBinary); err != nil {
		fmt.Fprintf(os.Stderr, "exo binary not found at %s\n", exoBinary)
		fmt.Fprintf(os.Stderr, "Please build the binary first: make build\n")
		os.Exit(1)
	}

	cmd := exec.Command(exoBinary, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}
	os.Exit(0)
}

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func stripANSI(s string) string { return ansiRe.ReplaceAllString(s, "") }

// ptyInput represents a single keystroke/sequence to be fed into a PTY process.
// If waitFor is non-empty, the input is held until that exact string appears
// somewhere in the accumulated PTY output (or a 10-second deadline expires),
// making test scenarios deterministic even on slow CI runners.
type ptyInput struct {
	waitFor string // pattern to wait for in PTY output before writing
	data    []byte // bytes to write to the PTY
}

// runInPTY starts cmd inside a PTY, optionally feeds keystrokes via the
// inputs channel, collects all PTY output with ANSI stripped, waits for the
// process to exit and returns the cleaned output.
//
// Each ptyInput can carry a waitFor pattern: when set, the byte sequence is
// not written until that string appears in the accumulated PTY output (or a
// 10-second timeout elapses). Inputs without a waitFor are still preceded by
// a short fixed delay so that fast-typing scenarios remain stable.
func runInPTY(ts *testscript.TestScript, cmd *exec.Cmd, inputs <-chan ptyInput) string {
	ptm, err := pty.Start(cmd)
	ts.Check(err)

	// mu protects rawOut, which is the ANSI-stripped PTY output accumulated so
	// far. The input goroutine reads it to detect wait-for patterns; the output
	// goroutine writes it as bytes arrive from the PTY.
	var mu sync.Mutex
	var rawOut strings.Builder

	// Output collector: read the PTY with a raw byte loop so that partial lines
	// (e.g. prompts that do not end with '\n') are captured immediately.
	outCh := make(chan string, 1)
	go func() {
		var cleaned strings.Builder
		buf := make([]byte, 4096)
		for {
			n, rerr := ptm.Read(buf)
			if n > 0 {
				chunk := stripANSI(string(buf[:n]))
				mu.Lock()
				rawOut.WriteString(chunk)
				mu.Unlock()
				// Accumulate per-line trimmed text for the return value.
				for _, line := range strings.Split(chunk, "\n") {
					if t := strings.TrimSpace(line); t != "" {
						cleaned.WriteString(t + "\n")
					}
				}
			}
			if rerr != nil {
				break
			}
		}
		outCh <- cleaned.String()
	}()

	if inputs != nil {
		go func() {
			for inp := range inputs {
				if inp.waitFor != "" {
					// Block until the expected string appears in PTY output.
					deadline := time.Now().Add(10 * time.Second)
					for time.Now().Before(deadline) {
						mu.Lock()
						found := strings.Contains(rawOut.String(), inp.waitFor)
						mu.Unlock()
						if found {
							break
						}
						time.Sleep(50 * time.Millisecond)
					}
				}
				if _, werr := ptm.Write(inp.data); werr != nil && werr != io.ErrClosedPipe {
					return
				}
			}
		}()
	}

	_ = cmd.Wait()
	_ = ptm.Close()
	return <-outCh
}

// cmdExecPTY mirrors the built-in exec but runs the binary inside a PTY.
// Usage: exec-pty [ <binary> [args...] ] [ <stdin-file> ]
// The first bracket group is the command; the second names a testscript file
// containing input tokens, one per line.
func cmdExecPTY(ts *testscript.TestScript, neg bool, args []string) {
	_, groups := splitByBrackets(args)
	if len(groups) != 2 {
		ts.Fatalf("exec-pty: usage: exec-pty [ <binary> [args...] ] [ <stdin-file> ]")
	}
	cmdArgs := groups[0]
	if len(cmdArgs) == 0 {
		ts.Fatalf("exec-pty: no binary specified")
	}
	if len(groups[1]) != 1 {
		ts.Fatalf("exec-pty: stdin group must contain exactly one filename")
	}
	stdinFile := groups[1][0]

	bin, err := exec.LookPath(cmdArgs[0])
	ts.Check(err)

	var tokens []string
	for _, line := range strings.Split(strings.TrimRight(ts.ReadFile(stdinFile), "\n"), "\n") {
		if t := strings.TrimSpace(line); t != "" {
			tokens = append(tokens, t)
		}
	}

	// pendingWait is set by an @wait:<pattern> line; it is attached to the
	// very next input token so that the write is delayed until the pattern is
	// visible in the PTY output.
	var pendingWait string

	inputs := make(chan ptyInput, len(tokens))
	for _, token := range tokens {
		if after, ok := strings.CutPrefix(token, "@wait:"); ok {
			pendingWait = after
			continue
		}

		inp := ptyInput{waitFor: pendingWait}
		pendingWait = ""

		switch token {
		case "@down", "↓":
			inp.data = []byte{'\x1b', '[', 'B'}
		case "@up", "↑":
			inp.data = []byte{'\x1b', '[', 'A'}
		case "@right", "→":
			inp.data = []byte{'\x1b', '[', 'C'}
		case "@left", "←":
			inp.data = []byte{'\x1b', '[', 'D'}
		case "@enter", "↵":
			inp.data = []byte{'\r'}
		case "@ctrl+c", "^C":
			inp.data = []byte{'\x03'}
		case "@ctrl+d", "^D":
			inp.data = []byte{'\x04'}
		case "@escape", "⎋":
			inp.data = []byte{'\x1b'}
		default:
			text := token
			if strings.HasPrefix(token, `\`) {
				text = token[1:] // strip the escape prefix, treat rest as literal
			}
			inp.data = []byte(text + "\r")
		}
		inputs <- inp
	}
	close(inputs)

	cmd := exec.Command(bin, cmdArgs[1:]...)
	cmd.Dir = ts.Getenv("WORK")

	envMap := make(map[string]string)
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}
	// HOME and XDG_CONFIG_HOME are used by os.UserConfigDir() for config location
	if v := ts.Getenv("HOME"); v != "" {
		envMap["HOME"] = v
	}
	if v := ts.Getenv("XDG_CONFIG_HOME"); v != "" {
		envMap["XDG_CONFIG_HOME"] = v
	}
	for k, v := range envMap {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	out := runInPTY(ts, cmd, inputs)

	exitCode := cmd.ProcessState.ExitCode()
	if exitCode != 0 {
		if !neg {
			ts.Fatalf("exec-pty %s: exit code %d\noutput:\n%s", cmdArgs[0], exitCode, out)
		}
		_, _ = fmt.Fprint(ts.Stderr(), out)
		return
	}
	if neg {
		ts.Fatalf("exec-pty %s: unexpectedly succeeded\noutput:\n%s", cmdArgs[0], out)
	}

	_, _ = fmt.Fprint(ts.Stdout(), out)
}

// splitByBrackets splits args into a leading options slice and bracket-delimited
// groups. Each group is the content between a "[" and its matching "]".
// Leading args before the first "[" are returned separately as options.
func splitByBrackets(args []string) (opts []string, groups [][]string) {
	i := 0
	for i < len(args) && args[i] != "[" {
		opts = append(opts, args[i])
		i++
	}
	for i < len(args) {
		if args[i] != "[" {
			break
		}
		i++ // skip "["
		var group []string
		for i < len(args) && args[i] != "]" {
			group = append(group, args[i])
			i++
		}
		if i < len(args) {
			i++ // skip "]"
		}
		groups = append(groups, group)
	}
	return opts, groups
}
