package key

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type keyDisableCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"disable"`

	Key string `cli-arg:"#" cli-usage:"ID"`

	Zone v3.ZoneName `cli-short:"z" cli-flag:"zone" cli-usage:"key zone"`
}

func (c *keyDisableCmd) CmdAliases() []string { return nil }

func (c *keyDisableCmd) CmdShort() string {
	return "Disables a KMS Key."
}

func (c *keyDisableCmd) CmdLong() string {
	return "Disables a KMS Key."
}

func (c *keyDisableCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *keyDisableCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	if _, err := client.DisableKmsKey(ctx, v3.UUID(c.Key)); err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&KeyShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Key:                c.Key,
			Zone:               c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(keyCmd, &keyDisableCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
