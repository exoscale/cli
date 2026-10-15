package keystore

import (
	"errors"
	"slices"
	"strings"
	"testing"

	exocmd "github.com/exoscale/cli/cmd"
	v3 "github.com/exoscale/egoscale/v3"
)

func TestKeyStoreShowByID(t *testing.T) {
	api := newFakeKeyStoreAPI(t)
	api.addKeyStore("my-xks", v3.GetKeyStoreResponseStatusConnected)

	c := &KeyStoreShowCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), KeyStore: string(testKeyStoreID)}
	if err := c.CmdRun(nil, nil); err != nil {
		t.Fatalf("keystore show: %v", err)
	}
	if want := []string{"get"}; !slices.Equal(api.calls, want) {
		t.Errorf("expected API calls %v, got %v", want, api.calls)
	}

	c = &KeyStoreShowCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), KeyStore: unknownKeyStoreID}
	if err := c.CmdRun(nil, nil); !errors.Is(err, v3.ErrNotFound) {
		t.Errorf("expected not found for an unknown key store, got %v", err)
	}
}

func TestKeyStoreCommandsRejectInvalidID(t *testing.T) {
	api := newFakeKeyStoreAPI(t)
	settings := exocmd.DefaultCLICmdSettings()

	for name, err := range map[string]error{
		"get":        (&KeyStoreShowCmd{CliCommandSettings: settings, KeyStore: "my-xks"}).CmdRun(nil, nil),
		"update":     (&keyStoreUpdateCmd{CliCommandSettings: settings, KeyStore: "my-xks", Description: "d"}).CmdRun(nil, nil),
		"connect":    (&keyStoreConnectCmd{CliCommandSettings: settings, KeyStore: "my-xks"}).CmdRun(nil, nil),
		"disconnect": (&keyStoreDisconnectCmd{CliCommandSettings: settings, KeyStore: "my-xks"}).CmdRun(nil, nil),
		"delete":     (&keyStoreDeleteCmd{CliCommandSettings: settings, KeyStore: "my-xks", Force: true}).CmdRun(nil, nil),
	} {
		if err == nil || !strings.Contains(err.Error(), "invalid key store ID") {
			t.Errorf("%s: expected an invalid key store ID error, got %v", name, err)
		}
	}

	if len(api.calls) != 0 {
		t.Fatalf("expected no API calls, got %v", api.calls)
	}
}

func TestNewKeyStoreShowOutput(t *testing.T) {
	ks := &v3.GetKeyStoreResponse{
		ID:     testKeyStoreID,
		Name:   "my-xks",
		Type:   v3.GetKeyStoreResponseTypeExternalKeyStore,
		Status: v3.GetKeyStoreResponseStatusConnected,
		Proxy: &v3.KeyStoreProxyResponse{
			Endpoint: "https://xks.example.com",
			Auth:     &v3.KeyStoreProxyAuthResponse{Key: "AKIDEXAMPLE"},
		},
		Health: &v3.KeyStoreHealth{
			Status:       v3.KeyStoreHealthStatusUnhealthy,
			StatusReason: "proxy-unreachable",
		},
	}

	out := newKeyStoreShowOutput(ks)
	if out.Endpoint != "https://xks.example.com" || out.AccessKey != "AKIDEXAMPLE" {
		t.Errorf("unexpected proxy fields: %q/%q", out.Endpoint, out.AccessKey)
	}
	if out.KeyStoreType != "external-key-store" || out.Status != "connected" {
		t.Errorf("unexpected type/status: %q/%q", out.KeyStoreType, out.Status)
	}
	if out.Health != "unhealthy" || out.HealthReason != "proxy-unreachable" {
		t.Errorf("unexpected health fields: %q/%q", out.Health, out.HealthReason)
	}

	// A key store that has not been health-checked yet is returned without health details.
	ks.Health = nil
	out = newKeyStoreShowOutput(ks)
	if out.Endpoint != "https://xks.example.com" {
		t.Errorf("unexpected endpoint: %q", out.Endpoint)
	}
	if out.Health != "" || out.HealthReason != "" || !out.HealthCheckedAt.IsZero() {
		t.Errorf("expected empty health fields, got %+v", out)
	}
}
