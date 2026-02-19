package e2e_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

var (
	exoBinary string
	cliRoot   string // Set at init time before working directory changes
)

func init() {
	// Use runtime.Caller to get the actual source file location
	// This works regardless of current working directory
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get source file location")
	}
	// filename is .../cli/internal/integ/testscript_test.go
	// We need .../cli
	cliRoot = filepath.Dir(filepath.Dir(filepath.Dir(filename)))

	// Use pre-built binary from the existing build pipeline
	// Tests should run against the actual build artifact, not rebuild it
	exoBinary = filepath.Join(cliRoot, "bin", "exo")
}

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"exo": mainExo,
	}))
}

func mainExo() int {
	// Check if binary exists
	if _, err := os.Stat(exoBinary); err != nil {
		fmt.Fprintf(os.Stderr, "exo binary not found at %s\n", exoBinary)
		fmt.Fprintf(os.Stderr, "Please build the binary first: make build\n")
		return 1
	}

	// Run the pre-built binary
	cmd := exec.Command(exoBinary, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		return 1
	}
	return 0
}
