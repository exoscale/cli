package instance

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceSnapshotCreateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Instance string `cli-arg:"#" cli-usage:"INSTANCE-NAME|ID"`

	Zone string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceSnapshotCreateCmd) CmdAliases() []string { return exocmd.GCreateAlias }

func (c *instanceSnapshotCreateCmd) CmdShort() string { return "Create a Compute instance snapshot" }

func (c *instanceSnapshotCreateCmd) CmdLong() string {
	return fmt.Sprintf(`This command creates a Compute instance snapshot.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instanceSnapshotShowOutput{}), ", "))
}

func (c *instanceSnapshotCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceSnapshotCreateCmd) CmdRun(_ *cobra.Command, _ []string) error {
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

	op, err := client.CreateSnapshot(ctx, instance.ID)
	if err != nil {
		return err
	}
	utils.DecorateAsyncOperation(fmt.Sprintf("Creating snapshot of instance %q...", c.Instance), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				err = fmt.Errorf("request timeout reached. Snapshot creation is not canceled and might still be running, check the status with: exo c i snapshot list") // nolint:stylecheck
			}
		}
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&instanceSnapshotShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			ID:                 op.Reference.ID.String(),
			Zone:               c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceSnapshotCmd, &instanceSnapshotCreateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
