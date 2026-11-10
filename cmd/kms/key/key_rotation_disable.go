package key

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type keyRotationDisableCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"disable-rotation"`

	Key string `cli-arg:"#" cli-usage:"ID"`

	Zone v3.ZoneName `cli-short:"z" cli-flag:"zone" cli-usage:"key zone"`
}

func (c *keyRotationDisableCmd) CmdAliases() []string { return nil }

func (c *keyRotationDisableCmd) CmdShort() string {
	return "Disable the periodic rotation of a KMS Key."
}

func (c *keyRotationDisableCmd) CmdLong() string {
	return "Disable the periodic rotation of a KMS Key."
}

func (c *keyRotationDisableCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	clicmd.CmdSetZoneFlagFromDefault(cmd)
	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *keyRotationDisableCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	if _, err := client.DisableKmsKeyRotation(ctx, v3.UUID(c.Key)); err != nil {
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
	cobra.CheckErr(clicmd.RegisterCLICommand(keyCmd, &keyRotationDisableCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
