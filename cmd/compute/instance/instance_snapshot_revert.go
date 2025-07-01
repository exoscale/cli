package instance

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceSnapshotRevertCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"revert"`

	Instance   string `cli-arg:"#" cli-usage:"INSTANCE-NAME|ID"`
	SnapshotID string `cli-arg:"#"`

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
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceSnapshotRevertCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	instances, err := client.ListInstances(ctx)
	if err != nil {
		return err
	}
	instance, err := instances.FindListInstancesResponseInstances(c.Instance)
	if err != nil {
		return err
	}

	snapshots, err := client.ListSnapshots(ctx)
	if err != nil {
		return err
	}
	snapshot, err := snapshots.FindSnapshot(c.SnapshotID)
	if err != nil {
		return err
	}

	if snapshot.Instance.ID != instance.ID {
		return fmt.Errorf("snapshot %s is not a snapshot of instance %s", snapshot.ID, instance.ID)
	}

	if !c.Force {
		if !utils.AskQuestion(
			ctx,
			fmt.Sprintf(
				"Are you sure you want to revert instance %q to snapshot %s?",
				c.Instance,
				c.SnapshotID,
			)) {
			return nil
		}
	}

	op, err := client.RevertInstanceToSnapshot(ctx, instance.ID, v3.RevertInstanceToSnapshotRequest{
		ID: snapshot.ID,
	})
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf(
		"Reverting instance %q to snapshot %s...",
		c.Instance,
		c.SnapshotID,
	), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&instanceShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Instance:           instance.ID.String(),
			Zone:               v3.ZoneName(c.Zone),
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceSnapshotCmd, &instanceSnapshotRevertCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
