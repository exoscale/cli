package cmd

import (
	"fmt"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type dbaasExternalEndpointDeleteCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Type       string `cli-arg:"#"`
	EndpointID string `cli-arg:"#"`
}

func (c *dbaasExternalEndpointDeleteCmd) CmdAliases() []string {
	return GDeleteAlias
}

func (c *dbaasExternalEndpointDeleteCmd) CmdLong() string {
	return "Delete a DBaaS external endpoint"
}

func (c *dbaasExternalEndpointDeleteCmd) CmdShort() string {
	return "Delete a DBaaS external endpoint"
}

func (c *dbaasExternalEndpointDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasExternalEndpointDeleteCmd) CmdRun(cmd *cobra.Command, args []string) error {

	ctx := GContext

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	endpointUUID, err := v3.ParseUUID(c.EndpointID)
	if err != nil {
		return fmt.Errorf("invalid endpoint ID: %w", err)
	}

	var op *v3.Operation
	var errOp error
	switch c.Type {
	case "datadog":
		op, errOp = client.DeleteDBAASExternalEndpointDatadog(ctx, endpointUUID)
	case "opensearch":
		op, errOp = client.DeleteDBAASExternalEndpointOpensearch(ctx, endpointUUID)
	case "elasticsearch":
		op, errOp = client.DeleteDBAASExternalEndpointElasticsearch(ctx, endpointUUID)
	case "prometheus":
		op, errOp = client.DeleteDBAASExternalEndpointPrometheus(ctx, endpointUUID)
	case "rsyslog":
		op, errOp = client.DeleteDBAASExternalEndpointRsyslog(ctx, endpointUUID)
	default:
		return fmt.Errorf("unsupported external endpoint type %q", c.Type)
	}

	if errOp != nil {
		return errOp
	}

	decorateAsyncOperation(fmt.Sprintf("Deleting external endpoint %s %s", c.Type, endpointUUID), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	return err
}

func init() {
	cobra.CheckErr(RegisterCLICommand(dbaasExternalEndpointCmd, &dbaasExternalEndpointDeleteCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
