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
	"sync/atomic"
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

// runInPTY starts cmd inside a PTY, optionally feeds keystrokes via the
// inputs channel, collects all PTY output with ANSI stripped, waits for the
// process to exit and returns the cleaned output.
//
// Instead of a fixed delay between keystrokes, it uses an output-settle
// approach: each keystroke is sent only once the PTY has produced output
// after the previous write and then gone quiet for settleDelay. This is
// similar to how `expect` works and is much more robust on slow CI runners.
func runInPTY(ts *testscript.TestScript, cmd *exec.Cmd, inputs <-chan []byte) string {
	ptm, err := pty.Start(cmd)
	ts.Check(err)

	// lastWrite is the UnixNano time of the most recent write to the PTY (0 = none).
	// lastActivity is the UnixNano time of the most recent byte read from the PTY (0 = none).
	// The input goroutine waits until lastActivity > lastWrite (output appeared after
	// the last write) and time.Since(lastActivity) >= settleDelay (output has gone quiet),
	// ensuring the prompt is ready before each keystroke is sent.
	const settleDelay = 80 * time.Millisecond
	var lastWrite, lastActivity atomic.Int64

	outCh := make(chan string, 1)
	go func() {
		var sb strings.Builder
		buf := make([]byte, 4096)
		for {
			n, rerr := ptm.Read(buf)
			if n > 0 {
				lastActivity.Store(time.Now().UnixNano())
				for _, line := range strings.Split(stripANSI(string(buf[:n])), "\n") {
					if line = strings.TrimSpace(line); line != "" {
						sb.WriteString(line + "\n")
					}
				}
			}
			if rerr != nil {
				break
			}
		}
		outCh <- sb.String()
	}()

	if inputs != nil {
		go func() {
			for b := range inputs {
				// Wait until output has appeared after the last write and settled.
				for {
					lw := lastWrite.Load()
					la := lastActivity.Load()
					if la > lw && time.Since(time.Unix(0, la)) >= settleDelay {
						break
					}
					time.Sleep(10 * time.Millisecond)
				}
				lastWrite.Store(time.Now().UnixNano())
				if _, werr := ptm.Write(b); werr != nil && werr != io.ErrClosedPipe {
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
// The input file is named explicitly via --stdin=<file>, removing any
// ambiguity with arguments forwarded to the binary itself.
func cmdExecPTY(ts *testscript.TestScript, neg bool, args []string) {
	var stdinFile string
	rest := args
	for i, a := range args {
		var found bool
		if stdinFile, found = strings.CutPrefix(a, "--stdin="); found {
			rest = append(args[:i:i], args[i+1:]...)
			break
		}
	}
	if stdinFile == "" {
		ts.Fatalf("execpty: usage: execpty --stdin=<file> <binary> [args...]")
	}
	if len(rest) == 0 {
		ts.Fatalf("execpty: no binary specified")
	}

	bin, err := exec.LookPath(rest[0])
	ts.Check(err)

	var tokens []string
	for _, line := range strings.Split(strings.TrimRight(ts.ReadFile(stdinFile), "\n"), "\n") {
		if t := strings.TrimSpace(line); t != "" {
			tokens = append(tokens, t)
		}
	}

	inputs := make(chan []byte, len(tokens))
	for _, token := range tokens {
		switch token {
		case "@down", "↓":
			inputs <- []byte{'\x1b', '[', 'B'}
		case "@up", "↑":
			inputs <- []byte{'\x1b', '[', 'A'}
		case "@right", "→":
			inputs <- []byte{'\x1b', '[', 'C'}
		case "@left", "←":
			inputs <- []byte{'\x1b', '[', 'D'}
		case "@enter", "↵":
			inputs <- []byte{'\r'}
		case "@ctrl+c", "^C":
			inputs <- []byte{'\x03'}
		case "@ctrl+d", "^D":
			inputs <- []byte{'\x04'}
		case "@escape", "⎋":
			inputs <- []byte{'\x1b'}
		default:
			text := token
			if strings.HasPrefix(token, `\`) {
				text = token[1:] // strip the escape prefix, treat rest as literal
			}
			inputs <- []byte(text + "\r")
		}
	}
	close(inputs)

	cmd := exec.Command(bin, rest[1:]...)
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
			ts.Fatalf("execpty %s: exit code %d\noutput:\n%s", rest[0], exitCode, out)
		}
		_, _ = fmt.Fprint(ts.Stderr(), out)
		return
	}
	if neg {
		ts.Fatalf("execpty %s: unexpectedly succeeded\noutput:\n%s", rest[0], out)
	}

	_, _ = fmt.Fprint(ts.Stdout(), out)
}
