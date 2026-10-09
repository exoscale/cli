package keystore

import (
	"errors"
	"slices"
	"testing"

	exocmd "github.com/exoscale/cli/cmd"
	v3 "github.com/exoscale/egoscale/v3"
)

func TestKeyStoreConnectDisconnect(t *testing.T) {
	api := newFakeKeyStoreAPI(t)
	ks := api.addKeyStore("my-xks", v3.GetKeyStoreResponseStatusDisconnected)

	connect := &keyStoreConnectCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), KeyStore: string(testKeyStoreID)}
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
	connect = &keyStoreConnectCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), KeyStore: unknownKeyStoreID}
	if err := connect.CmdRun(nil, nil); !errors.Is(err, v3.ErrNotFound) {
		t.Errorf("expected not found for an unknown key store, got %v", err)
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

	// --force skips the confirmation prompt.
	c := &keyStoreDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), KeyStore: string(testKeyStoreID), Force: true}
	if err := c.CmdRun(nil, nil); err != nil {
		t.Fatalf("keystore delete: %v", err)
	}
	if _, ok := api.stores[testKeyStoreID]; ok {
		t.Error("expected the key store to be deleted")
	}
	if want := []string{"delete"}; !slices.Equal(api.calls, want) {
		t.Errorf("expected API calls %v, got %v", want, api.calls)
	}

	c = &keyStoreDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), KeyStore: unknownKeyStoreID, Force: true}
	if err := c.CmdRun(nil, nil); !errors.Is(err, v3.ErrNotFound) {
		t.Errorf("expected not found for an unknown key store, got %v", err)
	}
}
