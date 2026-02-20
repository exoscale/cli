package e2e_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

// TestScriptsLocal runs testscript scenarios that don't require API access.
// These tests run by default without any build tags.
func TestScriptsLocal(t *testing.T) {
	// Find all txtar files recursively in scenarios/local
	files, err := findTestScripts("scenarios/local")
	if err != nil {
		t.Fatal(err)
	}

	testscript.Run(t, testscript.Params{
		Files: files,
		Setup: func(e *testscript.Env) error {
			return setupTestEnv(e, false)
		},
	})
}

// findTestScripts recursively finds all .txtar and .txt files in a directory
func findTestScripts(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (filepath.Ext(path) == ".txtar" || filepath.Ext(path) == ".txt") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// setupTestEnv configures the test environment for testscript scenarios.
// withAPI controls whether API credentials should be forwarded.
func setupTestEnv(e *testscript.Env, withAPI bool) error {
	// Redirect config directory to test's temp directory
	// This isolates config file changes per test
	configDir := filepath.Join(e.WorkDir, ".config")
	e.Setenv("XDG_CONFIG_HOME", configDir)
	e.Setenv("HOME", e.WorkDir)

	// Set default flags that all tests need
	// TODO: Make these parametrizable per test scenario
	e.Setenv("EXO_ZONE", "ch-gva-2")
	e.Setenv("EXO_OUTPUT", "json")

	// Forward API credentials if requested (for integration tests)
	if withAPI {
		if apiKey := os.Getenv("EXOSCALE_API_KEY"); apiKey != "" {
			e.Setenv("EXOSCALE_API_KEY", apiKey)
		}
		if apiSecret := os.Getenv("EXOSCALE_API_SECRET"); apiSecret != "" {
			e.Setenv("EXOSCALE_API_SECRET", apiSecret)
		}
	}

	return nil
}
