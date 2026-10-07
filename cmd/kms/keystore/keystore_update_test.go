package keystore

import (
	"testing"

	exocmd "github.com/exoscale/cli/cmd"
	v3 "github.com/exoscale/egoscale/v3"
)

func TestKeyStoreUpdateValidation(t *testing.T) {
	api := newFakeKeyStoreAPI(t)
	api.addKeyStore("my-xks", v3.GetKeyStoreResponseStatusConnected)

	for name, c := range map[string]*keyStoreUpdateCmd{
		"nothing to update":     {KeyStore: "my-xks"},
		"access key alone":      {KeyStore: "my-xks", AccessKey: "ak"},
		"secret key alone":      {KeyStore: "my-xks", SecretKey: "sk"},
		"unknown key store ref": {KeyStore: "unknown", Description: "d"},
	} {
		c.CliCommandSettings = exocmd.DefaultCLICmdSettings()
		if err := c.CmdRun(nil, nil); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}

	if api.called("update") {
		t.Fatal("update must not be sent for invalid input")
	}
}

func TestKeyStoreUpdate(t *testing.T) {
	api := newFakeKeyStoreAPI(t)
	api.addKeyStore("my-xks", v3.GetKeyStoreResponseStatusConnected)

	c := &keyStoreUpdateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		KeyStore:           "my-xks",
		Description:        "new description",
	}
	if err := c.CmdRun(nil, nil); err != nil {
		t.Fatalf("keystore update description: %v", err)
	}
	if api.updateReq.Description != "new description" || api.updateReq.Proxy != nil {
		t.Errorf("unexpected update request: %+v", api.updateReq)
	}

	c = &keyStoreUpdateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		KeyStore:           string(testKeyStoreID),
		Endpoint:           "https://xks2.example.com",
		AccessKey:          "AKIDNEW",
		SecretKey:          "newsecret",
	}
	if err := c.CmdRun(nil, nil); err != nil {
		t.Fatalf("keystore update proxy: %v", err)
	}
	proxy := api.updateReq.Proxy
	if proxy == nil || proxy.Endpoint != "https://xks2.example.com" || proxy.Auth == nil ||
		proxy.Auth.Key != "AKIDNEW" || proxy.Auth.Secret != "newsecret" {
		t.Errorf("unexpected proxy update: %+v", proxy)
	}

	c = &keyStoreUpdateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		KeyStore:           "my-xks",
		Endpoint:           "https://xks3.example.com",
	}
	if err := c.CmdRun(nil, nil); err != nil {
		t.Fatalf("keystore update endpoint only: %v", err)
	}
	if proxy := api.updateReq.Proxy; proxy == nil || proxy.Auth != nil {
		t.Errorf("endpoint-only update must not send credentials: %+v", proxy)
	}
}
