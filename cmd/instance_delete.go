package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instanceDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceDeleteCmd) cmdAliases() []string { return gRemoveAlias }

func (c *instanceDeleteCmd) cmdShort() string { return "Delete a Compute instance" }

func (c *instanceDeleteCmd) cmdLong() string { return "" }

func (c *instanceDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceDeleteCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	instance, err := globalstate.GlobalEgoscaleClient.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete instance %q?", c.Instance)) {
			return nil
		}
	}

	decorateAsyncOperation(fmt.Sprintf("Deleting instance %q...", c.Instance), func() {
		err = globalstate.GlobalEgoscaleClient.DeleteInstance(ctx, c.Zone, instance)
	})
	if err != nil {
		return err
	}

	instanceDir := path.Join(gConfigFolder, "instances", *instance.ID)
	if _, err := os.Stat(instanceDir); !os.IsNotExist(err) {
		if err := os.RemoveAll(instanceDir); err != nil {
			return fmt.Errorf("error deleting instance directory: %w", err)
		}
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceCmd, &instanceDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
