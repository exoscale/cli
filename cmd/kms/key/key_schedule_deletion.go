package key

import (
	"strconv"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type keyScheduleDeletionCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"schedule-deletion"`

	Key string `cli-arg:"#" cli-usage:"ID"`

	DelayDays string      `cli-short:"d" cli-flag:"delay-days" cli-usage:"number of days before deletion (7 - 30, default: 30)"`
	Zone      v3.ZoneName `cli-short:"z" cli-flag:"zone" cli-usage:"key zone"`
}

func (c *keyScheduleDeletionCmd) CmdAliases() []string { return nil }

func (c *keyScheduleDeletionCmd) CmdShort() string {
	return "Schedule a KMS key for deletion after a delay."
}

func (c *keyScheduleDeletionCmd) CmdLong() string {
	return "Schedule a KMS key for deletion after a delay."
}

func (c *keyScheduleDeletionCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *keyScheduleDeletionCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	var req v3.ScheduleKmsKeyDeletionRequest
	if cmd.Flags().Changed("delay-days") {
		n, err := strconv.Atoi(c.DelayDays)
		if err != nil {
			return err
		}
		req.DelayDays = n
	}

	_, err = client.ScheduleKmsKeyDeletion(ctx, v3.UUID(c.Key), req)
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
	cobra.CheckErr(exocmd.RegisterCLICommand(keyCmd, &keyScheduleDeletionCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
