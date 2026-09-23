package apikey

import (
	"context"
	"testing"
	"time"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

func TestAIAPIKeyGet(t *testing.T) {
	ts := newAPIKeyHelperServer(t)
	defer ts.server.Close()
	now := time.Now()
	ts.keys = []v3.AIAPIKey{
		{ID: v3.UUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), Name: "key1", Scope: "public", CreatedAT: now, UpdatedAT: now},
	}

	exocmd.GContext = context.Background()
	globalstate.Quiet = true
	creds := credentials.NewStaticCredentials("key", "secret")
	client, err := v3.NewClient(creds)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	globalstate.EgoscaleV3Client = client.WithEndpoint(v3.Endpoint(ts.server.URL))

	c := &AIAPIKeyGetCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), ID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}
	if err := c.CmdRun(nil, nil); err != nil {
		t.Fatalf("get api key: %v", err)
	}
}

func TestAIAPIKeyGetMissingID(t *testing.T) {
	ts := newAPIKeyHelperServer(t)
	defer ts.server.Close()

	exocmd.GContext = context.Background()
	globalstate.Quiet = true
	creds := credentials.NewStaticCredentials("key", "secret")
	client, err := v3.NewClient(creds)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	globalstate.EgoscaleV3Client = client.WithEndpoint(v3.Endpoint(ts.server.URL))

	c := &AIAPIKeyGetCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}
	if err := c.CmdRun(nil, nil); err == nil {
		t.Fatalf("expected error for missing ID")
	}
}
