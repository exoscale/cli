package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instanceSnapshotRevertCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"revert"`

	SnapshotID string `cli-arg:"#"`
	Instance   string `cli-arg:"#" cli-usage:"INSTANCE-NAME|ID"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"snapshot zone"`
}

func (c *instanceSnapshotRevertCmd) cmdAliases() []string { return nil }

func (c *instanceSnapshotRevertCmd) cmdShort() string {
	return "Revert a Compute instance to a snapshot"
}

func (c *instanceSnapshotRevertCmd) cmdLong() string {
	return fmt.Sprintf(`This command reverts a Compute instance to a snapshot.

/!\ **************************************************************** /!\
THIS OPERATION EFFECTIVELY RESTORES AN INSTANCE'S DISK TO A PREVIOUS
STATE: ALL DATA WRITTEN AFTER THE SNAPSHOT HAS BEEN CREATED WILL BE
LOST.
/!\ **************************************************************** /!\

Supported output template annotations: %s`,
		strings.Join(output.OutputterTemplateAnnotations(&instanceShowOutput{}), ", "))
}

func (c *instanceSnapshotRevertCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceSnapshotRevertCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	instance, err := globalstate.GlobalEgoscaleClient.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	snapshot, err := globalstate.GlobalEgoscaleClient.GetSnapshot(ctx, c.Zone, c.SnapshotID)
	if err != nil {
		return err
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf(
			"Are you sure you want to revert instance %q to snapshot %s?",
			c.Instance,
			c.SnapshotID,
		)) {
			return nil
		}
	}

	decorateAsyncOperation(fmt.Sprintf(
		"Reverting instance %q to snapshot %s...",
		c.Instance,
		c.SnapshotID,
	), func() {
		err = globalstate.GlobalEgoscaleClient.RevertInstanceToSnapshot(ctx, c.Zone, instance, snapshot)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&instanceShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Instance:           *instance.ID,
			Zone:               c.Zone,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceSnapshotCmd, &instanceSnapshotRevertCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
