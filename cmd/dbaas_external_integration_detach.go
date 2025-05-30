package cmd

import (
	"fmt"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type dbaasExternalIntegrationDetachCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"detach"`

	SourceServiceName string `cli-arg:"#"`

	IntegrationID string `cli-flag:"integration-id" cli-usage:"External integration id"`
}

func (c *dbaasExternalIntegrationDetachCmd) CmdAliases() []string {
	return []string{"a"}
}

func (c *dbaasExternalIntegrationDetachCmd) CmdLong() string {
	return "Disable sending data from an existing DBaaS service to an external endpoint"
}

func (c *dbaasExternalIntegrationDetachCmd) CmdShort() string {
	return "Detach a DBaaS service from an external endpoint"
}

func (c *dbaasExternalIntegrationDetachCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasExternalIntegrationDetachCmd) CmdRun(cmd *cobra.Command, args []string) error {

	ctx := GContext

	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	integrationID, err := v3.ParseUUID(c.IntegrationID)
	if err != nil {
		return fmt.Errorf("invalid integration ID: %w", err)
	}

	req := v3.DetachDBAASServiceFromEndpointRequest{
		IntegrationID: integrationID,
	}

	op, err := client.DetachDBAASServiceFromEndpoint(ctx, c.SourceServiceName, req)

	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Detaching service %s from endpoint %s", c.SourceServiceName, integrationID), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	return err
}

func init() {
	cobra.CheckErr(RegisterCLICommand(dbaasExternalIntegrationCmd, &dbaasExternalIntegrationDetachCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
