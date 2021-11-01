package cmd

import (
	"fmt"
	"strings"

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
		strings.Join(outputterTemplateAnnotations(&instanceShowOutput{}), ", "))
}

func (c *instanceSnapshotRevertCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceSnapshotRevertCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	instance, err := cs.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		return err
	}

	snapshot, err := cs.GetSnapshot(ctx, c.Zone, c.SnapshotID)
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
		err = cs.RevertInstanceToSnapshot(ctx, c.Zone, instance, snapshot)
	})
	if err != nil {
		return err
	}

	if !gQuiet {
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
