package instance

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type instanceScaleCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"scale"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`
	Type     string `cli-arg:"#" cli-usage:"SIZE"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceScaleCmd) CmdAliases() []string { return nil }

func (c *instanceScaleCmd) CmdShort() string { return "Scale a Compute instance" }

func (c *instanceScaleCmd) CmdLong() string {
	return fmt.Sprintf(`This commands scales a Compute instance to a different size.

Supported Compute instance type sizes: %s

Supported output template annotations: %s`,
		strings.Join(instanceTypeSizes, ", "),
		strings.Join(output.TemplateAnnotations(&InstanceShowOutput{}), ", "))
}

func (c *instanceScaleCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceScaleCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	instance, err := globalstate.EgoscaleClient.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	if !c.Force {
		if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to scale instance %q?", c.Instance)) {
			return nil
		}
	}

	instanceType, err := globalstate.EgoscaleClient.FindInstanceType(ctx, c.Zone, c.Type)
	if err != nil {
		return fmt.Errorf("error retrieving instance type: %w", err)
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Scaling instance %q...", c.Instance), func() {
		err = globalstate.EgoscaleClient.ScaleInstance(ctx, c.Zone, instance, instanceType)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&instanceShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Instance:           *instance.ID,
			Zone:               c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceCmd, &instanceScaleCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
