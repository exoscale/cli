package integ_local_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

var Binary = "../../../bin/exo"

// TestConfigPanic tests that config commands handle gracefully when
// defaultAccount field is missing from the config file.
//
// Note: With the fix, `exo config add` now automatically sets the first account
// as default, so this "no default" state should not occur through normal CLI usage.
// However, this state can still occur if:
// - User manually edits the config file and removes defaultAccount
// - Config file was created by an external tool
//
// This test ensures commands fail gracefully (not panic) when defaultAccount is missing.
//
// Note: This is an integration test rather than a testscript (e2e) test because
// `exo config add` uses interactive prompts (promptui) for zone selection and
// account information, which cannot be properly simulated in testscript's
// non-interactive environment.
func TestConfigPanic(t *testing.T) {
	tmpHome := t.TempDir()
	tmpConfigDir := filepath.Join(tmpHome, ".config", "exoscale")
	err := os.MkdirAll(tmpConfigDir, 0755)
	require.NoError(t, err)

	// Create config: account without defaultAccount field
	// This state can occur from manual config editing or external tools
	configWithoutDefault := `[[accounts]]
name = "test-account"
key = "EXOtest123"
secret = "testsecret"
defaultZone = "ch-gva-2"
`
	configPath := filepath.Join(tmpConfigDir, "exoscale.toml")
	err = os.WriteFile(configPath, []byte(configWithoutDefault), 0644)
	require.NoError(t, err)

	t.Run("commands handle missing default account gracefully", func(t *testing.T) {
		cmd := exec.Command(Binary, "config", "show")
		cmd.Env = append(os.Environ(), "HOME="+tmpHome)
		output, err := cmd.CombinedOutput()

		// Should fail gracefully with clear error message, not panic
		require.Error(t, err)
		require.Contains(t, string(output), "default account not defined")
		t.Logf("Output: %s", output)
	})

	t.Run("use-account flag bypasses default account requirement", func(t *testing.T) {
		cmd := exec.Command(Binary, "--use-account", "test-account", "config", "show")
		cmd.Env = append(os.Environ(), "HOME="+tmpHome)
		output, err := cmd.CombinedOutput()

		// Should work with --use-account flag
		if err != nil {
			// May fail due to invalid credentials, but shouldn't panic
			t.Logf("Command failed (expected with test credentials): %s", output)
		} else {
			require.Contains(t, string(output), "test-account")
		}
	})
}
