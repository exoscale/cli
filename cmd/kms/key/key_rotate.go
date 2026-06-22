package key

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type keyRotateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"rotate"`

	Key string `cli-arg:"#" cli-usage:"ID"`

	Zone v3.ZoneName `cli-short:"z" cli-flag:"zone" cli-usage:"key zone"`
}

func (c *keyRotateCmd) CmdAliases() []string { return nil }

func (c *keyRotateCmd) CmdShort() string {
	return "Perform a manual rotation of the key material for a symmetric key."
}

func (c *keyRotateCmd) CmdLong() string {
	return "Perform a manual rotation of the key material for a symmetric key."
}

func (c *keyRotateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *keyRotateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	_, err = client.RotateKmsKey(ctx, v3.UUID(c.Key))
	if err != nil {
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
	cobra.CheckErr(exocmd.RegisterCLICommand(keyCmd, &keyRotateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
