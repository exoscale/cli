//go:build api

package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
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
// Run with: go test -v -tags=api -timeout 10m
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
			"execpty":     cmdExecPTY,
			"exec-wait":   cmdExecWait,
			"json-setenv": cmdJSONSetenv,
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

// buildPollEnv returns a copy of the process environment with the
// testscript-isolated credentials and config directory substituted in, so that
// poll subprocesses use the same identity as the rest of the scenario.
// It also prepends the directory containing the exo binary to PATH so that
// commands in exec-wait brackets can reference "exo" by name.
func buildPollEnv(ts *testscript.TestScript) []string {
	overrides := map[string]string{
		"HOME":                ts.Getenv("HOME"),
		"XDG_CONFIG_HOME":     ts.Getenv("XDG_CONFIG_HOME"),
		"EXOSCALE_API_KEY":    ts.Getenv("EXOSCALE_API_KEY"),
		"EXOSCALE_API_SECRET": ts.Getenv("EXOSCALE_API_SECRET"),
	}
	env := make([]string, 0, len(os.Environ()))
	for _, kv := range os.Environ() {
		key := kv
		if i := strings.IndexByte(kv, '='); i >= 0 {
			key = kv[:i]
		}
		if _, overridden := overrides[key]; !overridden {
			env = append(env, kv)
		}
	}
	for k, v := range overrides {
		env = append(env, k+"="+v)
	}
	// Prepend the exo binary directory so "exo" resolves in exec-wait bracket commands.
	exoDir := filepath.Dir(exoBinary)
	for i, kv := range env {
		if strings.HasPrefix(kv, "PATH=") {
			env[i] = "PATH=" + exoDir + ":" + kv[5:]
			return env
		}
	}
	env = append(env, "PATH="+exoDir)
	return env
}

// cmdExecWait is a testscript custom command:
//
//	exec-wait [set=VARNAME:jsonfield ...] [ cmd1... ] [ cmd2... ] [ selector... ]
//
// Runs cmd1 once, optionally extracts JSON fields from its output into
// testscript env vars (set=) and builds {VARNAME} substitutions for cmd2.
// Then polls cmd2 every 10 seconds, piping its stdout into the selector process.
// The selector is any program that reads stdin and exits 0 when the condition
// is met (e.g. `jq -e '.state == "running"'`, `grep -q running`).
// Polling stops as soon as the selector exits 0.
//
// In cmd2 args, {VARNAME} tokens are replaced with values extracted by set=
// after cmd1 runs, allowing cmd2 to reference IDs not yet known at parse time.
func cmdExecWait(ts *testscript.TestScript, neg bool, args []string) {
	leadingOpts, groups := splitByBrackets(args)
	if len(groups) != 3 {
		ts.Fatalf("usage: exec-wait [set=VARNAME:jsonfield ...] [ cmd1... ] [ cmd2... ] [ selector... ]")
	}
	cmd1Args, cmd2Template, selectorArgs := groups[0], groups[1], groups[2]

	type setVar struct{ varName, jsonField string }
	var setVars []setVar

	for _, opt := range leadingOpts {
		if !strings.HasPrefix(opt, "set=") {
			ts.Fatalf("exec-wait: unknown option %q (only set= is allowed before first [)", opt)
		}
		kv := strings.TrimPrefix(opt, "set=")
		i := strings.IndexByte(kv, ':')
		if i < 0 {
			ts.Fatalf("exec-wait: invalid set= option %q, expected set=VARNAME:jsonfield", opt)
		}
		setVars = append(setVars, setVar{kv[:i], kv[i+1:]})
	}

	pollEnv := buildPollEnv(ts)

	// Run cmd1 once, capturing stdout only (stderr may contain spinner text).
	c1 := exec.Command(cmd1Args[0], cmd1Args[1:]...)
	c1.Env = pollEnv
	var c1Stderr bytes.Buffer
	c1.Stderr = &c1Stderr
	c1Stdout, c1Err := c1.Output()
	out := strings.TrimSpace(string(c1Stdout))
	if c1Err != nil {
		ts.Fatalf("exec-wait: cmd1 failed: %v\nstderr: %s", c1Err, c1Stderr.String())
	}

	// Extract set= vars and build {PLACEHOLDER} → value map.
	replacements := make(map[string]string, len(setVars))
	for _, sv := range setVars {
		val, err := parseJSONField(out, sv.jsonField)
		if err != nil {
			ts.Fatalf("exec-wait: set=%s:%s: %v", sv.varName, sv.jsonField, err)
		}
		ts.Setenv(sv.varName, val)
		replacements["{"+sv.varName+"}"] = val
	}

	// Resolve {PLACEHOLDER} tokens in the cmd2 template.
	cmd2 := make([]string, len(cmd2Template))
	for i, arg := range cmd2Template {
		resolved := arg
		for placeholder, val := range replacements {
			resolved = strings.ReplaceAll(resolved, placeholder, val)
		}
		cmd2[i] = resolved
	}

	// Poll: run cmd2 (stdout only), pipe into selector, stop when selector exits 0.
	for {
		c2 := exec.Command(cmd2[0], cmd2[1:]...)
		c2.Env = pollEnv
		var c2Stderr bytes.Buffer
		c2.Stderr = &c2Stderr
		c2Stdout, c2Err := c2.Output()
		if c2Err != nil {
			ts.Logf("exec-wait: cmd2 error (will retry): %v\nstderr: %s", c2Err, c2Stderr.String())
			time.Sleep(10 * time.Second)
			continue
		}

		cmd2Out := strings.TrimSpace(string(c2Stdout))
		sel := exec.Command(selectorArgs[0], selectorArgs[1:]...)
		sel.Stdin = bytes.NewBufferString(cmd2Out)
		selOut, selErr := sel.CombinedOutput()
		ts.Logf("exec-wait: selector output: %s", strings.TrimSpace(string(selOut)))
		if selErr == nil {
			return
		}
		ts.Logf("exec-wait: selector not satisfied (will retry): %v", selErr)
		time.Sleep(10 * time.Second)
	}
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

// runCLI runs the exo binary with the given arguments and returns combined stdout+stderr.
func runCLI(args ...string) (string, error) {
	return runCLIWithEnv(nil, args...)
}

// runCLIWithEnv runs the exo binary with an explicit environment (nil inherits the process env).
func runCLIWithEnv(env []string, args ...string) (string, error) {
	cmd := exec.Command(exoBinary, args...)
	cmd.Env = env
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
