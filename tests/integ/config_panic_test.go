package integ_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

var Binary = "../../bin/exo"

// TestConfigPanic tests the bug where adding a first account without setting
// it as default causes a panic at cmd/config/config.go:160
//
// Note: This is an integration test rather than a testscript (e2e) test because
// `exo config add` uses interactive prompts (promptui) for zone selection and
// account information, which cannot be properly simulated in testscript's
// non-interactive environment. Instead, we test the broken state that results
// from the panic - a config file with accounts but no defaultAccount field.
func TestConfigPanic(t *testing.T) {
	tmpHome := t.TempDir()
	tmpConfigDir := filepath.Join(tmpHome, ".config", "exoscale")
	err := os.MkdirAll(tmpConfigDir, 0755)
	require.NoError(t, err)

	// Create broken config: account without defaultAccount field
	brokenConfig := `[[accounts]]
name = "test-account"
key = "EXOtest123"
secret = "testsecret"
defaultZone = "ch-gva-2"
`
	configPath := filepath.Join(tmpConfigDir, "exoscale.toml")
	err = os.WriteFile(configPath, []byte(brokenConfig), 0644)
	require.NoError(t, err)

	t.Run("commands fail with broken config", func(t *testing.T) {
		cmd := exec.Command(Binary, "config", "show")
		cmd.Env = append(os.Environ(), "HOME="+tmpHome)
		output, err := cmd.CombinedOutput()

		require.Error(t, err)
		require.Contains(t, string(output), "default account not defined")
		t.Logf("Output: %s", output)
	})
}
