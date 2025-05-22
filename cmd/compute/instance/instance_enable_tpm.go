package instance

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceEnableTPMCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"enable-tpm"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`

	Force bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  v3.ZoneName `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceEnableTPMCmd) CmdAliases() []string { return nil }

func (c *instanceEnableTPMCmd) CmdShort() string { return "Enable Trusted Platform Module (TPM)" }

func (c *instanceEnableTPMCmd) CmdLong() string { return "" }

func (c *instanceEnableTPMCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceEnableTPMCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := GContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	resp, err := client.ListInstances(ctx)
	if err != nil {
		return err
	}

	instance, err := resp.FindListInstancesResponseInstances(c.Instance)
	if err != nil {
		return err
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to enable TPM on instance %q?", c.Instance)) {
			return nil
		}
	}

	op, err := client.EnableTpm(ctx, instance.ID)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Enabling Trusted Platform Module %q ...", c.Instance), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	return err
}

func init() {
	cobra.CheckErr(RegisterCLICommand(instanceCmd, &instanceEnableTPMCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
