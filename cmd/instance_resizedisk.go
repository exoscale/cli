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
)

type instanceResizeDiskCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"resize-disk"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`
	Size     int64  `cli-arg:"#"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceResizeDiskCmd) cmdAliases() []string { return nil }

func (c *instanceResizeDiskCmd) cmdShort() string { return "Resize a Compute instance disk" }

func (c *instanceResizeDiskCmd) cmdLong() string {
	return fmt.Sprintf(`This commands grows a Compute instance's disk to a larger size.'

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instanceShowOutput{}), ", "))
}

func (c *instanceResizeDiskCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceResizeDiskCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	instance, err := globalstate.EgoscaleClient.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to resize the disk of instance %q?", c.Instance)) {
			return nil
		}
	}

	decorateAsyncOperation(fmt.Sprintf("Resizing disk of instance %q...", c.Instance), func() {
		err = globalstate.EgoscaleClient.ResizeInstanceDisk(ctx, c.Zone, instance, c.Size)
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
	cobra.CheckErr(registerCLICommand(instanceCmd, &instanceResizeDiskCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
