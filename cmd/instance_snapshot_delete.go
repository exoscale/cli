package cmd

import (
	"fmt"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type computeInstanceSnapshotDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	ID string `cli-arg:"#"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"snapshot zone"`
}

func (c *computeInstanceSnapshotDeleteCmd) cmdAliases() []string { return gRemoveAlias }

func (c *computeInstanceSnapshotDeleteCmd) cmdShort() string {
	return "Delete a Compute instance snapshot"
}

func (c *computeInstanceSnapshotDeleteCmd) cmdLong() string { return "" }

func (c *computeInstanceSnapshotDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *computeInstanceSnapshotDeleteCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	snapshot, err := cs.GetSnapshot(ctx, c.Zone, c.ID)
	if err != nil {
		return err
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete snapshot %s?", c.ID)) {
			return nil
		}
	}

	decorateAsyncOperation(fmt.Sprintf("Deleting snapshot %s...", c.ID), func() {
		err = cs.DeleteSnapshot(ctx, c.Zone, *snapshot.ID)
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(computeInstanceSnapshotCmd, &computeInstanceSnapshotDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
