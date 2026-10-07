package keystore

import (
	"testing"

	exocmd "github.com/exoscale/cli/cmd"
	v3 "github.com/exoscale/egoscale/v3"
)

func TestKeyStoreShowByNameOrID(t *testing.T) {
	api := newFakeKeyStoreAPI(t)
	api.addKeyStore("my-xks", v3.GetKeyStoreResponseStatusConnected)

	for _, ref := range []string{"my-xks", string(testKeyStoreID)} {
		c := &KeyStoreShowCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), KeyStore: ref}
		if err := c.CmdRun(nil, nil); err != nil {
			t.Fatalf("keystore get %s: %v", ref, err)
		}
	}

	c := &KeyStoreShowCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), KeyStore: "unknown"}
	if err := c.CmdRun(nil, nil); err == nil {
		t.Fatal("expected an error for an unknown key store")
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
