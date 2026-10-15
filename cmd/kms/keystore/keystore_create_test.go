package keystore

import (
	"testing"

	exocmd "github.com/exoscale/cli/cmd"
	v3 "github.com/exoscale/egoscale/v3"
)

func TestKeyStoreCreateRequiresProxyFlags(t *testing.T) {
	api := newFakeKeyStoreAPI(t)

	for name, c := range map[string]*keyStoreCreateCmd{
		"missing access key": {Name: "ks", Endpoint: "https://xks.example.com", SecretKey: "sk"},
		"missing secret key": {Name: "ks", Endpoint: "https://xks.example.com", AccessKey: "ak"},
	} {
		c.CliCommandSettings = exocmd.DefaultCLICmdSettings()
		if err := c.CmdRun(nil, nil); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}

	if len(api.calls) != 0 {
		t.Fatalf("expected no API calls, got %v", api.calls)
	}
}

func TestKeyStoreCreate(t *testing.T) {
	api := newFakeKeyStoreAPI(t)

	c := &keyStoreCreateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		Name:               "my-xks",
		Description:        "test store",
		Endpoint:           "https://xks.example.com",
		AccessKey:          "AKIDEXAMPLE",
		SecretKey:          "secret",
	}
	if err := c.CmdRun(nil, nil); err != nil {
		t.Fatalf("keystore create: %v", err)
	}

	req := api.createReq
	if req == nil {
		t.Fatal("create request not sent")
	}
	if req.Name != "my-xks" || req.Description != "test store" {
		t.Errorf("unexpected name/description: %q/%q", req.Name, req.Description)
	}
	if req.Type != v3.CreateKeyStoreRequestTypeExternalKeyStore {
		t.Errorf("unexpected type %q", req.Type)
	}
	if req.Proxy.Endpoint != "https://xks.example.com" || req.Proxy.Auth.Key != "AKIDEXAMPLE" || req.Proxy.Auth.Secret != "secret" {
		t.Errorf("unexpected proxy settings: %+v / %+v", req.Proxy, req.Proxy.Auth)
	}
	if api.called("connect") {
		t.Error("connect must not be called without --connect-on-create")
	}
	if got := api.stores[testKeyStoreID].Status; got != v3.GetKeyStoreResponseStatusDisconnected {
		t.Errorf("expected disconnected key store, got %q", got)
	}
}

func TestKeyStoreCreateConnectOnCreate(t *testing.T) {
	api := newFakeKeyStoreAPI(t)

	c := &keyStoreCreateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		Name:               "my-xks",
		Endpoint:           "https://xks.example.com",
		AccessKey:          "AKIDEXAMPLE",
		SecretKey:          "secret",
		ConnectOnCreate:    true,
	}
	if err := c.CmdRun(nil, nil); err != nil {
		t.Fatalf("keystore create --connect-on-create: %v", err)
	}

	if !api.called("create") || !api.called("connect") {
		t.Fatalf("expected create then connect, got %v", api.calls)
	}
	if got := api.stores[testKeyStoreID].Status; got != v3.GetKeyStoreResponseStatusConnected {
		t.Errorf("expected connected key store, got %q", got)
	}
}
