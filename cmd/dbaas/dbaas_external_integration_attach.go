package dbaas

import (
	"fmt"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type dbaasExternalIntegrationAttachCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"attach"`

	Type string `cli-arg:"#"`

	SourceServiceName     string `cli-flag:"source-service-name" cli-usage:"DBaaS source service name"`
	DestinationEndpointID string `cli-flag:"destination-endpoint-id" cli-usage:"Destination external endpoint id"`
}

func (c *dbaasExternalIntegrationAttachCmd) CmdAliases() []string {
	return []string{"a"}
}

func (c *dbaasExternalIntegrationAttachCmd) CmdLong() string {
	return "Enable sending data from an existing DBaaS service to an external endpoint"
}

func (c *dbaasExternalIntegrationAttachCmd) CmdShort() string {
	return "Attach a DBaaS service to an external endpoint"
}

func (c *dbaasExternalIntegrationAttachCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasExternalIntegrationAttachCmd) CmdRun(cmd *cobra.Command, args []string) error {

	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	endpointID, err := v3.ParseUUID(c.DestinationEndpointID)
	if err != nil {
		return fmt.Errorf("invalid endpoint ID: %w", err)
	}

	req := v3.AttachDBAASServiceToEndpointRequest{
		DestEndpointID: endpointID,
		Type:           v3.EnumExternalEndpointTypes(c.Type),
	}

	op, err := client.AttachDBAASServiceToEndpoint(ctx, c.SourceServiceName, req)

	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Attaching service %s to endpoint %s", c.SourceServiceName, endpointID), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	return err
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasExternalIntegrationCmd, &dbaasExternalIntegrationAttachCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
