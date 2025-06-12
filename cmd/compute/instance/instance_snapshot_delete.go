package instance

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
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
	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	snapshot, err := globalstate.EgoscaleClient.GetSnapshot(ctx, c.Zone, c.ID)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	if !c.Force {
		if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to delete snapshot %s?", c.ID)) {
			return nil
		}
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Deleting snapshot %s...", c.ID), func() {
		err = globalstate.EgoscaleClient.DeleteSnapshot(ctx, c.Zone, snapshot)
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
