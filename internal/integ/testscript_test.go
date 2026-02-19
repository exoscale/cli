package integ_test

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
			// TODO: Make these parametrizable per test scenario
			e.Setenv("EXO_ZONE", "ch-gva-2")
			e.Setenv("EXO_OUTPUT", "json")

			// Forward API credentials from environment (for CI and local API tests)
			// Tests can use either env vars or config files
			// TODO: Currently no scenarios use API credentials. Future PR will add
			//       API-based tests requiring org account setup with proper credentials.
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
