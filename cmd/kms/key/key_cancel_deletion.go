package key

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type keyCancelDeletionCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"cancel-deletion"`

	Key string `cli-arg:"#" cli-usage:"ID"`

	Zone v3.ZoneName `cli-short:"z" cli-flag:"zone" cli-usage:"key zone"`
}

func (c *keyCancelDeletionCmd) CmdAliases() []string { return nil }

func (c *keyCancelDeletionCmd) CmdShort() string {
	return "Cancel the scheduled deletion of a KMS Key."
}

func (c *keyCancelDeletionCmd) CmdLong() string {
	return "Cancel the scheduled deletion of a KMS Key."
}

func (c *keyCancelDeletionCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *keyCancelDeletionCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	_, err = client.CancelKmsKeyDeletion(ctx, v3.UUID(c.Key))
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&KeyShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Key:                c.Key,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(keyCmd, &keyCancelDeletionCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
