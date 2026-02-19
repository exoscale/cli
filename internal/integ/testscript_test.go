package integ_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

var (
	buildOnce  sync.Once
	buildError error
	exoBinary  string
	cliRoot    string // Set at init time before working directory changes
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
}

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"exo": mainExo,
	}))
}

// buildExoBinary builds the exo CLI binary once for all tests
func buildExoBinary() error {
	buildOnce.Do(func() {
		// Verify go.mod exists
		if _, err := os.Stat(filepath.Join(cliRoot, "go.mod")); err != nil {
			buildError = fmt.Errorf("go.mod not found in %s: %w", cliRoot, err)
			return
		}

		// Create bin directory if it doesn't exist
		binDir := filepath.Join(cliRoot, "bin")
		if err := os.MkdirAll(binDir, 0755); err != nil {
			buildError = fmt.Errorf("failed to create bin directory: %w", err)
			return
		}

		exoBinary = filepath.Join(binDir, "exo")

		// Build the binary - use the main.go path explicitly
		mainPath := filepath.Join(cliRoot, "main.go")
		cmd := exec.Command("go", "build", "-o", exoBinary, mainPath)
		cmd.Dir = cliRoot
		output, err := cmd.CombinedOutput()
		if err != nil {
			buildError = fmt.Errorf("failed to build exo binary: %w\n%s", err, output)
			return
		}
	})
	return buildError
}

func mainExo() int {
	// Build binary once if not already built
	if err := buildExoBinary(); err != nil {
		fmt.Fprintf(os.Stderr, "build error: %v\n", err)
		return 1
	}

	// Run the compiled binary
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

func TestScripts(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "scenarios",
		Setup: func(e *testscript.Env) error {
			// Redirect config directory to test's temp directory
			// This isolates config file changes per test
			configDir := filepath.Join(e.WorkDir, ".config")
			e.Setenv("XDG_CONFIG_HOME", configDir)
			e.Setenv("HOME", e.WorkDir)

			// Set default flags that all tests need
			e.Setenv("EXO_ZONE", "ch-gva-2")
			e.Setenv("EXO_OUTPUT", "json")

			// Forward API credentials from environment (for CI and local API tests)
			// Tests can use either env vars or config files
			if apiKey := os.Getenv("EXOSCALE_API_KEY"); apiKey != "" {
				e.Setenv("EXOSCALE_API_KEY", apiKey)
			}
			if apiSecret := os.Getenv("EXOSCALE_API_SECRET"); apiSecret != "" {
				e.Setenv("EXOSCALE_API_SECRET", apiSecret)
			}

			// Alternatively, copy real config for tests that need it:
			// (Uncomment if you prefer config file over env vars)
			// if realConfig, err := os.ReadFile(filepath.Join(os.Getenv("HOME"), ".config/exoscale/exoscale.toml")); err == nil {
			//     testConfigPath := filepath.Join(configDir, "exoscale", "exoscale.toml")
			//     os.MkdirAll(filepath.Dir(testConfigPath), 0755)
			//     os.WriteFile(testConfigPath, realConfig, 0644)
			// }

			return nil
		},
	})
}
