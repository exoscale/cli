package apikey

import (
	"context"
	"testing"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

func TestAIAPIKeyDeleteValidationAndSuccess(t *testing.T) {
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

	// missing IDs
	c := &AIAPIKeyDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}
	if err := c.CmdRun(nil, nil); err == nil {
		t.Fatalf("expected error for missing IDs")
	}

	// success with force
	c.IDs = []string{"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}
	c.Force = true
	if err := c.CmdRun(nil, nil); err != nil {
		t.Fatalf("delete api key: %v", err)
	}
}
