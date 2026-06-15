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

func TestAIAPIKeyList(t *testing.T) {
	ts := newAPIKeyHelperServer(t)
	defer ts.server.Close()
	now := time.Now()
	ts.keys = []v3.AIAPIKey{
		{ID: v3.UUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), Name: "key1", Scope: "public", CreatedAT: now, UpdatedAT: now},
		{ID: v3.UUID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"), Name: "key2", Scope: "public", CreatedAT: now, UpdatedAT: now},
	}

	exocmd.GContext = context.Background()
	globalstate.Quiet = true
	creds := credentials.NewStaticCredentials("key", "secret")
	client, err := v3.NewClient(creds)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	globalstate.EgoscaleV3Client = client.WithEndpoint(v3.Endpoint(ts.server.URL))

	cmd := &AIAPIKeyListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("list api keys: %v", err)
	}
}
