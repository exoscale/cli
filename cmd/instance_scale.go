package cmd

import (
	"errors"
	"fmt"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instanceScaleCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"scale"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`
	Type     string `cli-arg:"#" cli-usage:"SIZE"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceScaleCmd) cmdAliases() []string { return nil }

func (c *instanceScaleCmd) cmdShort() string { return "Scale a Compute instance" }

func (c *instanceScaleCmd) cmdLong() string {
	return fmt.Sprintf(`This commands scales a Compute instance to a different size.

Supported Compute instance type sizes: %s

Supported output template annotations: %s`,
		strings.Join(instanceTypeSizes, ", "),
		strings.Join(outputterTemplateAnnotations(&instanceShowOutput{}), ", "))
}

func (c *instanceScaleCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceScaleCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	instance, err := cs.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to scale instance %q?", c.Instance)) {
			return nil
		}
	}

	instanceType, err := cs.FindInstanceType(ctx, c.Zone, c.Type)
	if err != nil {
		return fmt.Errorf("error retrieving instance type: %w", err)
	}

	decorateAsyncOperation(fmt.Sprintf("Scaling instance %q...", c.Instance), func() {
		err = cs.ScaleInstance(ctx, c.Zone, instance, instanceType)
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
	cobra.CheckErr(registerCLICommand(instanceCmd, &instanceScaleCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
