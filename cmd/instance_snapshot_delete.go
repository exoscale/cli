package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type instanceSnapshotDeleteCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	ID string `cli-arg:"#"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"snapshot zone"`
}

func (c *instanceSnapshotDeleteCmd) CmdAliases() []string { return GRemoveAlias }

func (c *instanceSnapshotDeleteCmd) CmdShort() string {
	return "Delete a Compute instance snapshot"
}

func (c *instanceSnapshotDeleteCmd) CmdLong() string { return "" }

func (c *instanceSnapshotDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceSnapshotDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	snapshot, err := globalstate.EgoscaleClient.GetSnapshot(ctx, c.Zone, c.ID)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete snapshot %s?", c.ID)) {
			return nil
		}
	}

	decorateAsyncOperation(fmt.Sprintf("Deleting snapshot %s...", c.ID), func() {
		err = globalstate.EgoscaleClient.DeleteSnapshot(ctx, c.Zone, snapshot)
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(RegisterCLICommand(instanceSnapshotCmd, &instanceSnapshotDeleteCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
