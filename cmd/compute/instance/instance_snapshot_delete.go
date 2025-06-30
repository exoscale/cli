package instance

import (
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceSnapshotDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	ID string `cli-arg:"#"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"snapshot zone"`
}

func (c *instanceSnapshotDeleteCmd) CmdAliases() []string { return exocmd.GRemoveAlias }

func (c *instanceSnapshotDeleteCmd) CmdShort() string {
	return "Delete a Compute instance snapshot"
}

func (c *instanceSnapshotDeleteCmd) CmdLong() string { return "" }

func (c *instanceSnapshotDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceSnapshotDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	snapshots, err := client.ListSnapshots(ctx)
	if err != nil {
		return err
	}
	snapshot, err := snapshots.FindSnapshot(c.ID)
	if err != nil {
		return err
	}

	if !c.Force {
		if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to delete snapshot %s for instance %s?", snapshot.ID, snapshot.Instance.ID)) {
			return nil
		}
	}

	op, err := client.DeleteSnapshot(ctx, snapshot.ID)
	if err != nil {
		return err
	}
	utils.DecorateAsyncOperation(fmt.Sprintf("Deleting snapshot %s...", c.ID), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceSnapshotCmd, &instanceSnapshotDeleteCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
