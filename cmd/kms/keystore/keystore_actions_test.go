package keystore

import (
	"testing"

	exocmd "github.com/exoscale/cli/cmd"
	v3 "github.com/exoscale/egoscale/v3"
)

func TestKeyStoreConnectDisconnect(t *testing.T) {
	api := newFakeKeyStoreAPI(t)
	ks := api.addKeyStore("my-xks", v3.GetKeyStoreResponseStatusDisconnected)

	connect := &keyStoreConnectCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), KeyStore: "my-xks"}
	if err := connect.CmdRun(nil, nil); err != nil {
		t.Fatalf("keystore connect: %v", err)
	}
	if ks.Status != v3.GetKeyStoreResponseStatusConnected {
		t.Errorf("expected connected, got %q", ks.Status)
	}

	disconnect := &keyStoreDisconnectCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), KeyStore: string(testKeyStoreID)}
	if err := disconnect.CmdRun(nil, nil); err != nil {
		t.Fatalf("keystore disconnect: %v", err)
	}
	if ks.Status != v3.GetKeyStoreResponseStatusDisconnected {
		t.Errorf("expected disconnected, got %q", ks.Status)
	}

	connect = &keyStoreConnectCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), KeyStore: "unknown"}
	if err := connect.CmdRun(nil, nil); err == nil {
		t.Error("expected an error for an unknown key store")
	}
}

func TestKeyStoreList(t *testing.T) {
	api := newFakeKeyStoreAPI(t)
	api.addKeyStore("my-xks", v3.GetKeyStoreResponseStatusConnected)

	c := &keyStoreListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}
	if err := c.CmdRun(nil, nil); err != nil {
		t.Fatalf("keystore list: %v", err)
	}
	if !api.called("list") {
		t.Error("expected a list call")
	}
}

func TestKeyStoreDelete(t *testing.T) {
	api := newFakeKeyStoreAPI(t)
	api.addKeyStore("my-xks", v3.GetKeyStoreResponseStatusDisconnected)

	// Unknown key store without --force fails before deleting anything.
	c := &keyStoreDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), KeyStores: []string{"unknown"}}
	if err := c.CmdRun(nil, nil); err == nil {
		t.Fatal("expected an error for an unknown key store")
	}

	// With --force, unknown key stores are skipped and known ones deleted.
	c = &keyStoreDeleteCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		KeyStores:          []string{"unknown", "my-xks"},
		Force:              true,
	}
	if err := c.CmdRun(nil, nil); err != nil {
		t.Fatalf("keystore delete: %v", err)
	}
	if _, ok := api.stores[testKeyStoreID]; ok {
		t.Error("expected the key store to be deleted")
	}
}
