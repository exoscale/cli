package key

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type keyReplicateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"replicate"`

	Key        string      `cli-arg:"#" cli-usage:"ID"`
	TargetZone v3.ZoneName `cli-arg:"#" cli-usage:"TARGET_ZONE"`

	Zone v3.ZoneName `cli-short:"z" cli-flag:"zone" cli-usage:"key zone"`
}

func (c *keyReplicateCmd) CmdAliases() []string { return nil }

func (c *keyReplicateCmd) CmdShort() string {
	return "Replicate a KMS key to a target zone."
}

func (c *keyReplicateCmd) CmdLong() string {
	return "Replicate a KMS key to a target zone."
}

func (c *keyReplicateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *keyReplicateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	req := v3.ReplicateKmsKeyRequest{
		Zone: string(c.TargetZone),
	}

	resp, err := client.ReplicateKmsKey(ctx, v3.UUID(c.Key), req)
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		out := successResponseOutput{
			Status: resp.Status,
		}
		return c.OutputFunc(&out, nil)
	}
	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(keyCmd, &keyReplicateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
