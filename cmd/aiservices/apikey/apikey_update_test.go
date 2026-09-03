package apikey

import (
	"context"
	"testing"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

func TestAIAPIKeyUpdateValidationAndSuccess(t *testing.T) {
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

	// missing flags
	c := &AIAPIKeyUpdateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), ID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}
	if err := c.CmdRun(nil, nil); err == nil {
		t.Fatalf("expected error for missing flags")
	}

	// success
	c.Name = "updated-key"
	if err := c.CmdRun(nil, nil); err != nil {
		t.Fatalf("update api key: %v", err)
	}
}
