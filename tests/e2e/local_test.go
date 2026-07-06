//go:build local_integration

package e2e_test

// Local dev only — use -tags=local_integration -account=<name>

import (
	"flag"
	"os"
	"strings"
	"testing"
)

var flagAccount = flag.String("account", "owner-production", "account name substring in exoscale.toml")

func TestAPIStorageLocal(t *testing.T) {
	loadLocalCreds(t)
	runAPITestSuite(t, "scenarios/with-api/storage")
}

func TestAPIComputeLocal(t *testing.T) {
	loadLocalCreds(t)
	runAPITestSuite(t, "scenarios/with-api/compute")
}

func TestAPIDBaaSLocal(t *testing.T) {
	loadLocalCreds(t)
	runAPITestSuite(t, "scenarios/with-api/dbaas")
}

func loadLocalCreds(t *testing.T) {
	t.Helper()

	accountName := *flagAccount

	toml, err := os.ReadFile(os.ExpandEnv("$HOME/.config/exoscale/exoscale.toml"))
	if err != nil {
		t.Fatalf("loadLocalCreds: %v", err)
	}

	blocks := strings.Split(string(toml), "[[accounts]]")
	for _, block := range blocks[1:] {
		if !strings.Contains(block, accountName) {
			continue
		}
		key := tomlString(block, "key")
		secret := tomlString(block, "secret")
		zone := tomlString(block, "defaultZone")
		if key == "" || secret == "" {
			continue
		}
		t.Setenv("EXOSCALE_API_KEY", key)
		t.Setenv("EXOSCALE_API_SECRET", secret)
		if zone != "" {
			t.Setenv("EXOSCALE_ZONE", zone)
		}
		return
	}

	t.Fatalf("loadLocalCreds: no account matching %q in exoscale.toml", accountName)
}

func tomlString(block, key string) string {
	prefix := key + " = '"
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, prefix); ok {
			return strings.TrimSuffix(after, "'")
		}
	}
	return ""
}
