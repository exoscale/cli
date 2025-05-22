package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	exoapi "github.com/exoscale/egoscale/v2/api"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceSnapshotRevertCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"revert"`

	SnapshotID string `cli-arg:"#"`
	Instance   string `cli-arg:"#" cli-usage:"INSTANCE-NAME|ID"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"snapshot zone"`
}

func (c *instanceSnapshotRevertCmd) CmdAliases() []string { return nil }

func (c *instanceSnapshotRevertCmd) CmdShort() string {
	return "Revert a Compute instance to a snapshot"
}

func (c *instanceSnapshotRevertCmd) CmdLong() string {
	return fmt.Sprintf(`This command reverts a Compute instance to a snapshot.

/!\ **************************************************************** /!\
THIS OPERATION EFFECTIVELY RESTORES AN INSTANCE'S DISK TO A PREVIOUS
STATE: ALL DATA WRITTEN AFTER THE SNAPSHOT HAS BEEN CREATED WILL BE
LOST.
/!\ **************************************************************** /!\

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&InstanceShowOutput{}), ", "))
}

func (c *instanceSnapshotRevertCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceSnapshotRevertCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	instance, err := globalstate.EgoscaleClient.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	snapshot, err := globalstate.EgoscaleClient.GetSnapshot(ctx, c.Zone, c.SnapshotID)
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
		err = globalstate.EgoscaleClient.RevertInstanceToSnapshot(ctx, c.Zone, instance, snapshot)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&instanceShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Instance:           *instance.ID,
			Zone:               v3.ZoneName(c.Zone),
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(RegisterCLICommand(instanceSnapshotCmd, &instanceSnapshotRevertCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
