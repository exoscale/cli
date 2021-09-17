package cmd

import (
	"fmt"
	"strings"
	"time"

	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
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
		strings.Join(outputterTemplateAnnotations(&instanceSnapshotShowOutput{}), ", "))
}

func (c *instanceSnapshotCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceSnapshotCreateCmd) cmdRun(_ *cobra.Command, _ []string) error {
	// Snapshot creation can take a _long time_, raising
	// the Exoscale API client timeout as a precaution.
	cs.Client.SetTimeout(30 * time.Minute)

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	instance, err := cs.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		return err
	}

	var snapshot *egoscale.Snapshot
	decorateAsyncOperation(fmt.Sprintf("Creating snapshot of instance %q...", c.Instance), func() {
		snapshot, err = cs.CreateInstanceSnapshot(ctx, c.Zone, instance)
		if err != nil {
			return
		}
	})
	if err != nil {
		return err
	}

	if !gQuiet {
		return output(showInstanceSnapshot(c.Zone, *snapshot.ID))
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceSnapshotCmd, &instanceSnapshotCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
