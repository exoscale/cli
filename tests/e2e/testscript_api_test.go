//go:build api

package e2e_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/rogpeppe/go-internal/testscript"
)

// APITestSuite holds per-run metadata shared across all scenarios.
// Each scenario manages the lifecycle of its own resources.
type APITestSuite struct {
	Zone  string
	RunID string
}

// TestScriptsAPI runs testscript scenarios that require real API access.
// Run with: go test -v -tags=api -timeout 30m
//
// Required environment variables:
//
//	EXOSCALE_API_KEY    - Exoscale API key
//	EXOSCALE_API_SECRET - Exoscale API secret
//	EXOSCALE_ZONE       - Zone to run tests in (default: ch-gva-2)
func TestScriptsAPI(t *testing.T) {
	if os.Getenv("EXOSCALE_API_KEY") == "" || os.Getenv("EXOSCALE_API_SECRET") == "" {
		t.Skip("Skipping API tests: EXOSCALE_API_KEY and EXOSCALE_API_SECRET must be set")
	}

	zone := os.Getenv("EXOSCALE_ZONE")
	if zone == "" {
		zone = "ch-gva-2"
	}

	runID := fmt.Sprintf("e2e-%d-%s", time.Now().Unix(), randString(6))
	t.Logf("API test run ID: %s (zone: %s)", runID, zone)

	suite := &APITestSuite{
		Zone:  zone,
		RunID: runID,
	}

	// Run all scenarios under scenarios/with-api/
	files, err := findTestScripts("scenarios/with-api")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Log("No API test scenarios found in scenarios/with-api/")
		return
	}

	testscript.Run(t, testscript.Params{
		Files: files,
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"execpty":           cmdExecPTY,
			"json-setenv":       cmdJSONSetenv,
			"wait-instance-state": cmdWaitInstanceState,
		},
		Setup: func(e *testscript.Env) error {
			return setupAPITestEnv(e, suite)
		},
	})
}

// setupAPITestEnv configures the testscript environment with API credentials
// and run metadata. Each scenario creates and deletes its own resources.
func setupAPITestEnv(e *testscript.Env, suite *APITestSuite) error {
	// Isolate config directory
	e.Setenv("XDG_CONFIG_HOME", e.WorkDir+"/.config")
	e.Setenv("HOME", e.WorkDir)

	// API credentials
	e.Setenv("EXOSCALE_API_KEY", os.Getenv("EXOSCALE_API_KEY"))
	e.Setenv("EXOSCALE_API_SECRET", os.Getenv("EXOSCALE_API_SECRET"))

	// Zone and run metadata
	e.Setenv("EXO_ZONE", suite.Zone)
	e.Setenv("EXO_OUTPUT", "json")
	e.Setenv("TEST_RUN_ID", suite.RunID)
	e.Setenv("TEST_ZONE", suite.Zone)

	// Write a ready-to-use config file so scenarios don't need to run exo config add
	configDir := e.WorkDir + "/.config/exoscale"
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	configContent := fmt.Sprintf(`defaultAccount = "e2e-test"

[[accounts]]
name = "e2e-test"
key = "%s"
secret = "%s"
defaultZone = "%s"
`,
		os.Getenv("EXOSCALE_API_KEY"),
		os.Getenv("EXOSCALE_API_SECRET"),
		suite.Zone,
	)
	return os.WriteFile(configDir+"/exoscale.toml", []byte(configContent), 0600)
}

// cmdJSONSetenv is a testscript custom command:
//
//	json-setenv VARNAME FIELD FILE
//
// Reads FILE (relative to WorkDir), parses it as JSON, and sets the env var
// VARNAME to the string value of top-level FIELD. If the JSON is an array,
// the first element is used.
//
// Typical use after capturing exec output:
//
//	exec exo --output-format json compute instance create ... > out.json
//	json-setenv INSTANCE_ID id out.json
func cmdJSONSetenv(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) != 3 {
		ts.Fatalf("usage: json-setenv VARNAME FIELD FILE")
	}
	varName, field, file := args[0], args[1], args[2]
	content := ts.ReadFile(file)
	val, err := parseJSONField(content, field)
	if err != nil {
		ts.Fatalf("json-setenv: %v", err)
	}
	ts.Setenv(varName, val)
}

// cmdWaitInstanceState is a testscript custom command:
//
//	wait-instance-state ZONE INSTANCE_ID TARGET_STATE [TIMEOUT_SECONDS]
//
// Polls `exo compute instance show` until the instance reaches TARGET_STATE
// or the timeout elapses (default: 300 seconds). Fails the test on timeout.
func cmdWaitInstanceState(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) < 3 || len(args) > 4 {
		ts.Fatalf("usage: wait-instance-state ZONE INSTANCE_ID TARGET_STATE [TIMEOUT_SECONDS]")
	}
	zone, instanceID, targetState := args[0], args[1], args[2]
	timeout := 300 * time.Second
	if len(args) == 4 {
		var secs int
		if _, err := fmt.Sscan(args[3], &secs); err != nil {
			ts.Fatalf("wait-instance-state: invalid timeout %q: %v", args[3], err)
		}
		timeout = time.Duration(secs) * time.Second
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		out, err := runCLI(
			"--zone", zone,
			"--output-format", "json",
			"compute", "instance", "show",
			instanceID,
		)
		if err != nil {
			ts.Logf("wait-instance-state: poll error (will retry): %v", err)
			time.Sleep(10 * time.Second)
			continue
		}
		state, err := parseJSONField(out, "state")
		if err != nil {
			time.Sleep(10 * time.Second)
			continue
		}
		ts.Logf("wait-instance-state: %s → %s (want: %s)", instanceID, state, targetState)
		if state == targetState {
			return
		}
		time.Sleep(10 * time.Second)
	}
	ts.Fatalf("wait-instance-state: timed out waiting for instance %s to reach state %q", instanceID, targetState)
}

// runCLI runs the exo binary with the given arguments and returns combined stdout+stderr.
func runCLI(args ...string) (string, error) {
	cmd := exec.Command(exoBinary, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// parseJSONField extracts a top-level string field from a JSON object.
func parseJSONField(jsonStr, field string) (string, error) {
	// Handle JSON arrays by taking the first element
	trimmed := strings.TrimSpace(jsonStr)
	if strings.HasPrefix(trimmed, "[") {
		var arr []json.RawMessage
		if err := json.Unmarshal([]byte(trimmed), &arr); err != nil {
			return "", err
		}
		if len(arr) == 0 {
			return "", fmt.Errorf("empty JSON array")
		}
		trimmed = string(arr[0])
	}

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &obj); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}
	val, ok := obj[field]
	if !ok {
		return "", fmt.Errorf("field %q not found in JSON", field)
	}
	return fmt.Sprintf("%v", val), nil
}

func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}
