package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type instanceSnapshotCreateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Instance string `cli-arg:"#" cli-usage:"INSTANCE-NAME|ID"`

	Zone string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceSnapshotCreateCmd) cmdAliases() []string { return gCreateAlias }

func (c *instanceSnapshotCreateCmd) cmdShort() string { return "Create a Compute instance snapshot" }

func (c *instanceSnapshotCreateCmd) cmdLong() string {
	return fmt.Sprintf(`This command creates a Compute instance snapshot.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instanceSnapshotShowOutput{}), ", "))
}

func (c *instanceSnapshotCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceSnapshotCreateCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	instance, err := globalstate.EgoscaleClient.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		return err
	}

	var snapshot *egoscale.Snapshot
	decorateAsyncOperation(fmt.Sprintf("Creating snapshot of instance %q...", c.Instance), func() {
		snapshot, err = globalstate.EgoscaleClient.CreateInstanceSnapshot(ctx, c.Zone, instance)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				err = fmt.Errorf("request timeout reached. Snapshot creation is not canceled and might still be running, check the status with: exo c i snapshot list") // nolint:stylecheck
			}
			return
		}
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&instanceSnapshotShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			ID:                 *snapshot.ID,
			Zone:               c.Zone,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceSnapshotCmd, &instanceSnapshotCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
